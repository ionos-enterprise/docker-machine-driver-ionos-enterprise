package profitbricks

import (
	"errors"
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/profitbricks/profitbricks-sdk-go"
	"io/ioutil"
	"strings"
	"github.com/profitbricks/profitbricks-sdk-go/model"
)

type Driver struct {
	*drivers.BaseDriver
	URL          string
	Username     string
	Password     string
	ServerId     string
	Ram          int
	Cores        int
	SSHKey       string
	DatacenterId string
	DiskSize     int
	DiskType     string
	Image        string
	Size         int
	Location     string
}

const (
	defaultRegion = "us/lasdev"
	defaultSize = 10
	waitCount = 5
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_ENDPOINT",
			Name:   "profitbricks-endpoint",
			Value:  "https://api.profitbricks.com/rest/v2",
			Usage:  "profitbricks API endpoint",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_USERNAME",
			Name:   "profitbricks-username",
			Usage:  "profitbricks username",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_PASSWORD",
			Name:   "profitbricks-password",
			Usage:  "profitbricks password",
		},
		mcnflag.IntFlag{
			EnvVar: "PROFITBRICKS_CORES",
			Name:   "profitbricks-cores",
			Value:  4,
			Usage:  "profitbricks cores (2, 3, 4, 5, 6, etc.)",
		},
		mcnflag.IntFlag{
			EnvVar: "PROFITBRICKS_RAM",
			Name:   "profitbricks-ram",
			Value:  2048,
			Usage:  "profitbricks ram (1024, 2048, 3072, 4096, etc.)",
		},
		mcnflag.IntFlag{
			EnvVar: "PROFITBRICKS_DISK_SIZE",
			Name:   "profitbricks-disk-size",
			Value:  50,
			Usage:  "profitbricks disk size (10, 50, 100, 200, 400)",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_IMAGE",
			Name:   "profitbricks-image",
			Value:  "Ubuntu-16.04",
			Usage:  "profitbricks image",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_LOCATION",
			Name:   "profitbricks-location",
			Value:  "us/las",
			Usage:  "profitbricks location",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_DISK_TYPE",
			Name:   "profitbricks-disk-type",
			Value:  "HDD",
			Usage:  "profitbricks disk type (HDD, SSD)",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		Size:     defaultSize,
		Location: defaultRegion,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) DriverName() string {
	return "profitbricks"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {

	d.URL = flags.String("profitbricks-endpoint")
	d.Username = flags.String("profitbricks-username")

	d.Password = flags.String("profitbricks-password")
	d.DiskSize = flags.Int("profitbricks-disk-size")
	d.Image = flags.String("profitbricks-image")
	d.Cores = flags.Int("profitbricks-cores")
	d.Ram = flags.Int("profitbricks-ram")
	d.Location = flags.String("profitbricks-location")
	d.DiskType = flags.String("profitbricks-disk-type")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SetSwarmConfigFromFlags(flags)

	if d.URL == "" {
		d.URL = "https://api.profitbricks.com/rest/v2"
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.Username == "" {
		return fmt.Errorf("Please provide username as paramter --profitbricks-username or as environment variable $PROFITBRICKS_USERNAME")
	}
	if d.getImageId(d.Image) == "" {
		return fmt.Errorf("The image %s %s %s", d.Image, d.Location, "does not exist.")
	}
	return nil
}

func (d *Driver) Create() error {

	d.setPB()

	var err error
	if d.SSHKey == "" {
		d.SSHKey, err = d.createSSHKey()
		if err != nil {
			return err
		}
	}
	image := d.getImageId(d.Image)

	ipblockreq := profitbricks.IPBlockReserveRequest{
		IPBlockProperties: profitbricks.IPBlockProperties{
			Size:     1,
			Location: d.Location,
		},
	}

	ipblockresp := profitbricks.ReserveIpBlock(ipblockreq)

	d.waitTillProvisioned(strings.Join(ipblockresp.Resp.Headers["Location"], ""))

	datacenter := model.Datacenter{
		Properties: model.DatacenterProperties{
			Name: d.MachineName,
			Location:d.Location,
		},
		Entities:model.DatacenterEntities{
			Servers: &model.Servers{
				Items:[]model.Server{
					model.Server{
						Properties: model.ServerProperties{
							Name : d.MachineName,
							Ram: d.Ram,
							Cores: d.Cores,
						},
						Entities:model.ServerEntities{
							Volumes: &model.AttachedVolumes{
								Items:[]model.Volume{
									model.Volume{
										Properties: model.VolumeProperties{
											Type_:d.DiskType,
											Size:d.DiskSize,
											Name:d.MachineName,
											Image:image,
											SshKeys: []string{d.SSHKey},
										},
									},
								},
							},
							Nics: &model.Nics{
								Items: []model.Nic{
									model.Nic{
										Properties: model.NicProperties{
											Name : d.MachineName,
											Lan : "1",
											Ips: []string{ipblockresp.Properties["ips"].([]interface{})[0].(string)},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	dc := profitbricks.CompositeCreateDatacenter(datacenter)
	if dc.StatusCode == 202 {
		log.Info("Datacenter Created")
	} else {
		return errors.New("Error while creating DC: " + string(dc.Response))

	}

	d.DatacenterId = dc.Id

	d.waitTillProvisioned(dc.Headers.Get("Location"))

	lanrequest := profitbricks.CreateLanRequest{
		LanProperties: profitbricks.LanProperties{
			Public: true,
		},
	}

	lan := profitbricks.CreateLan(dc.Id, lanrequest)

	if lan.Resp.StatusCode == 202 {
		log.Info("LAN Created")
	} else {
		log.Error("Error while creating a LAN " + string(lan.Resp.Body))
		d.Remove()
		return errors.New("Rolling back...")
	}

	d.waitTillProvisioned(strings.Join(lan.Resp.Headers["Location"], ""))

	obj := profitbricks.PatchNic(dc.Id, dc.Entities.Servers.Items[0].Id, dc.Entities.Servers.Items[0].Entities.Nics.Items[0].Id, map[string]string{"lan": lan.Id})

	d.waitTillProvisioned(obj.Resp.Headers.Get("Location"))
	d.ServerId = dc.Entities.Servers.Items[0].Id

	d.IPAddress = ipblockresp.Properties["ips"].([]interface{})[0].(string)
	log.Info(d.IPAddress)
	return nil
}

func (d *Driver) Restart() error {
	d.setPB()
	resp := profitbricks.RebootServer(d.DatacenterId, d.ServerId)
	if resp.StatusCode != 202 {
		return errors.New(string(resp.Body))
	}
	return nil
}
func (d *Driver) Remove() error {
	d.setPB()

	resp := profitbricks.DeleteDatacenter(d.DatacenterId)
	d.waitTillProvisioned(strings.Join(resp.Headers["Location"], ""))
	ipblocks := profitbricks.ListIpBlocks()

	for i := 0; i < len(ipblocks.Items); i++ {
		for _, v := range ipblocks.Items[i].Properties["ips"].([]interface{}) {
			if d.IPAddress == v.(string) {
				resp := profitbricks.ReleaseIpBlock(ipblocks.Items[i].Id)
				if resp.StatusCode > 299 {
					return errors.New(string(resp.Body))
				}
			}
		}
	}

	if resp.StatusCode > 299 {
		return errors.New(string(resp.Body))
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) Start() error {
	serverstate, err := d.GetState()
	d.setPB()

	if err != nil {
		return err
	}
	if serverstate != state.Running {
		profitbricks.StartServer(d.DatacenterId, d.ServerId)
	} else {
		log.Info("Host is already running or starting")
	}
	return nil
}

func (d *Driver) Stop() error {
	vmstate, err := d.GetState()
	if err != nil {
		return err
	}
	if vmstate == state.Stopped {
		log.Infof("Host is already stopped")
		return nil
	}

	d.setPB()
	profitbricks.StopServer(d.DatacenterId, d.ServerId)
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) Kill() error {
	resp := profitbricks.StopServer(d.DatacenterId, d.ServerId)
	if resp.StatusCode != 202 {
		return errors.New(string(resp.Body))
	}
	return nil
}
func (d *Driver) GetIP() (string, error) {
	d.setPB()
	server := profitbricks.GetServer(d.DatacenterId, d.ServerId)

	d.IPAddress = server.Entities["nics"].Items[0].Properties["ips"].([]interface{})[0].(string)
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	d.setPB()
	server := profitbricks.GetServer(d.DatacenterId, d.ServerId)

	switch server.MetaData["state"] {
	case "NOSTATE":
		return state.None, nil
	case "AVAILABLE":
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case "BLOCKED":
		return state.Stopped, nil
	case "SHUTDOWN":
		return state.Stopped, nil
	case "SHUTOFF":
		return state.Stopped, nil
	case "CHRASHED":
		return state.Error, nil
	case "INACTIVE":
		return state.Stopped, nil
	}
	return state.None, nil
}

//Private helper functions
func (d *Driver) setPB() {
	profitbricks.SetAuth(d.Username, d.Password)
	profitbricks.SetEndpoint(d.URL)
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) createSSHKey() (string, error) {

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}
	return string(publicKey), nil
}

func (d *Driver) isSwarmMaster() bool {
	return d.SwarmMaster
}

func (d *Driver) waitTillProvisioned(path string) {
	d.setPB()
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.MetaData["status"] == "DONE" {
			break
		}
		time.Sleep(10 * time.Second)
		i++
	}
}

func (d *Driver) getImageId(imageName string) string {
	d.setPB()

	images := profitbricks.ListImages()

	if (images.Resp.StatusCode == 401) {
		log.Error("Authentication failed")
		return ""
	}

	for i := 0; i < len(images.Items); i++ {
		imgName := ""
		if images.Items[i].Properties["name"] != nil {
			imgName = images.Items[i].Properties["name"].(string)
		}
		diskType := d.DiskType
		if (d.DiskType == "SSD") {
			diskType = "HDD"
		}
		if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && images.Items[i].Properties["imageType"] == diskType && images.Items[i].Properties["location"] == d.Location {
			return images.Items[i].Id
		}
	}
	return ""
}
