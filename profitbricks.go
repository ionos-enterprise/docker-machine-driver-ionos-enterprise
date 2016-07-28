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
	"strconv"
	"strings"
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
	defaultSize   = 10
	waitCount     = 100
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

	ipblockreq := profitbricks.IpBlock{
		Properties: profitbricks.IpBlockProperties{
			Size:     1,
			Location: d.Location,
		},
	}

	ipblockresp := profitbricks.ReserveIpBlock(ipblockreq)

	if ipblockresp.StatusCode > 299 {
		return fmt.Errorf("An error occurred while reserving an ipblock: %s", ipblockresp.StatusCode)
	}

	d.waitTillProvisioned(ipblockresp.Headers.Get("Location"))

	datacenter := profitbricks.Datacenter{
		Properties: profitbricks.DatacenterProperties{
			Name:     d.MachineName,
			Location: d.Location,
		},
		Entities: profitbricks.DatacenterEntities{
			Servers: &profitbricks.Servers{
				Items: []profitbricks.Server{
					profitbricks.Server{
						Properties: profitbricks.ServerProperties{
							Name:  d.MachineName,
							Ram:   d.Ram,
							Cores: d.Cores,
						},
						Entities: &profitbricks.ServerEntities{
							Volumes: &profitbricks.Volumes{
								Items: []profitbricks.Volume{
									profitbricks.Volume{
										Properties: profitbricks.VolumeProperties{
											Type:    d.DiskType,
											Size:    d.DiskSize,
											Name:    d.MachineName,
											Image:   image,
											SshKeys: []string{d.SSHKey},
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
	d.waitTillProvisioned(dc.Headers.Get("Location"))

	lan := profitbricks.CreateLan(dc.Id, profitbricks.Lan{
		Properties: profitbricks.LanProperties{
			Public: true,
			Name:   d.MachineName,
		},
	})

	if lan.StatusCode == 202 {
		log.Info("LAN Created")
	} else {
		log.Error()
		d.Remove()
		return errors.New("Error while creating a LAN " + string(lan.Response) + "Rolling back...")
	}

	d.DatacenterId = dc.Id

	d.waitTillProvisioned(lan.Headers.Get("Location"))
	lanId, _ := strconv.Atoi(lan.Id)

	nic := profitbricks.CreateNic(dc.Id, dc.Entities.Servers.Items[0].Id, profitbricks.Nic{
		Properties: profitbricks.NicProperties{
			Name: d.MachineName,
			Lan:  lanId,
			Ips:  ipblockresp.Properties.Ips,
		},
	})

	if nic.StatusCode == 202 {
		log.Info("NIC Created")
	} else {
		d.Remove()
		return errors.New("Error while creating a NIC " + string(nic.Response) + "Rolling back...")
	}

	d.waitTillProvisioned(nic.Headers.Get("Location"))
	d.ServerId = dc.Entities.Servers.Items[0].Id

	d.IPAddress = ipblockresp.Properties.Ips[0]
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
		for _, v := range ipblocks.Items[i].Properties.Ips {
			if d.IPAddress == v {
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

	d.IPAddress = server.Entities.Nics.Items[0].Properties.Ips[0]
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	d.setPB()
	server := profitbricks.GetServer(d.DatacenterId, d.ServerId)

	switch server.Metadata.State {
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
		if request.Metadata.Status == "DONE" {
			break
		}
		time.Sleep(10 * time.Second)
		i++
	}
}

func (d *Driver) getImageId(imageName string) string {
	d.setPB()

	images := profitbricks.ListImages()

	if images.StatusCode == 401 {
		log.Error("Authentication failed")
		return ""
	}

	for i := 0; i < len(images.Items); i++ {
		imgName := ""
		if images.Items[i].Properties.Name != "" {
			imgName = images.Items[i].Properties.Name
		}
		diskType := d.DiskType
		if d.DiskType == "SSD" {
			diskType = "HDD"
		}
		if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && images.Items[i].Properties.ImageType == diskType && images.Items[i].Properties.Location == d.Location {
			return images.Items[i].Id
		}
	}
	return ""
}
