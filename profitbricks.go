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
	URL                    string
	Username               string
	Password               string
	ServerId               string
	Ram                    int
	Cores                  int
	SSHKey                 string
	DatacenterId           string
	VolumeAvailabilityZone string
	ServerAvailabilityZone string
	DiskSize               int
	DiskType               string
	Image                  string
	Size                   int
	Location               string
	CpuFamily              string
	DCExists               bool
	LanId                  string
}

const (
	defaultRegion = "us/las"
	defaultSize = 10
	waitCount = 1000
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_ENDPOINT",
			Name:   "profitbricks-endpoint",
			Value:  "https://api.profitbricks.com/cloudapi/v3",
			Usage:  "ProfitBricks API endpoint",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_USERNAME",
			Name:   "profitbricks-username",
			Usage:  "ProfitBricks username",
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
			Usage:  "ProfitBricks cores (2, 3, 4, 5, 6, etc.)",
		},
		mcnflag.IntFlag{
			EnvVar: "PROFITBRICKS_RAM",
			Name:   "profitbricks-ram",
			Value:  2048,
			Usage:  "ProfitBricks ram (1024, 2048, 3072, 4096, etc.)",
		},
		mcnflag.IntFlag{
			EnvVar: "PROFITBRICKS_DISK_SIZE",
			Name:   "profitbricks-disk-size",
			Value:  50,
			Usage:  "ProfitBricks disk size (10, 50, 100, 200, 400)",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_IMAGE",
			Name:   "profitbricks-image",
			Value:  "Ubuntu-16.04",
			Usage:  "ProfitBricks image",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_LOCATION",
			Name:   "profitbricks-location",
			Value:  "us/las",
			Usage:  "ProfitBricks location",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_DISK_TYPE",
			Name:   "profitbricks-disk-type",
			Value:  "HDD",
			Usage:  "ProfitBricks disk type (HDD, SSD)",
		},
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_CPU_FAMILY",
			Name:   "profitbricks-cpu-family",
			Value:  "AMD_OPTERON",
			Usage:  "ProfitBricks CPU families (AMD_OPTERON,INTEL_XEON)",
		},
		mcnflag.StringFlag{
			Name:   "profitbricks-datacenter-id",
			Usage:  "ProfitBricks Virtual Data Center Id",
		},
		mcnflag.StringFlag{
			Name:   "profitbricks-volume-availability-zone",
			Value: "AUTO",
			Usage:  "ProfitBricks Volume Availability Zone (AUTO, ZONE_1, ZONE_2, ZONE_3)",
		},
		mcnflag.StringFlag{
			Name:   "profitbricks-server-availability-zone",
			Value: "AUTO",
			Usage:  "ProfitBricks Server Availability Zone (AUTO, ZONE_1, ZONE_2, ZONE_3)",
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
	d.CpuFamily = flags.String("profitbricks-cpu-family")
	d.DatacenterId = flags.String("profitbricks-datacenter-id")
	d.VolumeAvailabilityZone = flags.String("profitbricks-volume-availability-zone")
	d.ServerAvailabilityZone = flags.String("profitbricks-server-availability-zone")

	d.SetSwarmConfigFromFlags(flags)

	if d.URL == "" {
		d.URL = "https://api.profitbricks.com/cloudapi/v3"
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.Username == "" {
		return fmt.Errorf("Please provide username as paramter --profitbricks-username or as environment variable $PROFITBRICKS_USERNAME")
	}
	if (d.DatacenterId != "") {
		d.setPB()

		dc := profitbricks.GetDatacenter(d.DatacenterId)

		if (dc.StatusCode == 404) {
			return fmt.Errorf("DataCenter UUID %s does not exist.", d.DatacenterId)
		} else {
			log.Info("Creating machine under " + dc.Properties.Name + " datacenter.")
		}
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

	err = d.waitTillProvisioned(ipblockresp.Headers.Get("Location"))
	if err != nil {
		return err
	}

	var dc profitbricks.Datacenter

	if (d.DatacenterId == "") {
		d.DCExists = false
		dc = profitbricks.Datacenter{
			Properties: profitbricks.DatacenterProperties{
				Name:     d.MachineName,
				Location: d.Location,
			},
		}

		dc = profitbricks.CompositeCreateDatacenter(dc)
		if dc.StatusCode == 202 {
			log.Info("Datacenter Created")
		} else {
			return errors.New("Error while creating DC: " + string(dc.Response))

		}
		err = d.waitTillProvisioned(dc.Headers.Get("Location"))
		if err != nil {
			return err
		}
	} else {
		d.DCExists = true
		dc = profitbricks.GetDatacenter(d.DatacenterId)
	}

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

	err = d.waitTillProvisioned(lan.Headers.Get("Location"))
	if err != nil {
		return err
	}

	lanId, _ := strconv.Atoi(lan.Id)

	d.LanId = lan.Id

	server := profitbricks.Server{
		Properties: profitbricks.ServerProperties{
			Name:  d.MachineName,
			Ram:   d.Ram,
			Cores: d.Cores,
			CpuFamily: d.CpuFamily,
			AvailabilityZone: d.ServerAvailabilityZone,
		},
		Entities: &profitbricks.ServerEntities{
			Volumes: &profitbricks.Volumes{
				Items: []profitbricks.Volume{
					{
						Properties: profitbricks.VolumeProperties{
							Type:    d.DiskType,
							Size:    d.DiskSize,
							Name:    d.MachineName,
							Image:   image,
							SshKeys: []string{d.SSHKey},
							AvailabilityZone: d.VolumeAvailabilityZone,
						},
					},
				},
			},
		},
	}

	nic := profitbricks.Nic{
		Properties: profitbricks.NicProperties{
			Name: d.MachineName,
			Lan:  lanId,
			Ips:  ipblockresp.Properties.Ips,
		},
	}

	server.Entities.Nics = &profitbricks.Nics{
		Items: []profitbricks.Nic{
			nic,
		},
	}

	server = profitbricks.CreateServer(dc.Id, server)

	if server.StatusCode == 202 {
		log.Info("Server Created")
	} else {
		d.Remove()
		return errors.New("Error while creating a server " + string(server.Response) + "Rolling back...")
	}

	err = d.waitTillProvisioned(server.Headers.Get("Location"))
	if err != nil {
		return err
	}
	d.ServerId = server.Id

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

	if (!d.DCExists) {
		servers := profitbricks.ListServers(d.DatacenterId)
		if (len(servers.Items) == 1) {
			resp := profitbricks.DeleteDatacenter(d.DatacenterId)
			if resp.StatusCode > 299 {
				return errors.New(string(resp.Body))
			}

			err := d.waitTillProvisioned(strings.Join(resp.Headers["Location"], ""))
			if err != nil {
				return err
			}
		} else {
			err := d.removeServer(d.DatacenterId, d.ServerId, d.LanId)
			if err != nil {
				return err
			}
		}
	} else {
		err := d.removeServer(d.DatacenterId, d.ServerId, d.LanId)
		if err != nil {
			return err
		}
	}

	ipblocks := profitbricks.ListIpBlocks()

	for _, i := range ipblocks.Items {
		for _, v := range i.Properties.Ips {
			if d.IPAddress == v {
				resp := profitbricks.ReleaseIpBlock(i.Id)
				if resp.StatusCode > 299 {
					return errors.New(string(resp.Body))
				}
			}
		}
	}

	return nil
}
func (d *Driver) removeServer(datacenterId string, serverId string, lanId string) error {
	server := profitbricks.GetServer(datacenterId, serverId)

	if server.StatusCode > 299 {
		return errors.New(server.Response)
	}

	if (server.Entities != nil && server.Entities.Volumes != nil && len(server.Entities.Volumes.Items) > 0) {
		volumeId := server.Entities.Volumes.Items[0].Id
		resp := profitbricks.DeleteVolume(d.DatacenterId, volumeId)
		if resp.StatusCode > 299 {
			return errors.New(string(resp.Body))
		}
		err := d.waitTillProvisioned(resp.Headers.Get("Location"))

		if err != nil {
			return err
		}
	}
	resp := profitbricks.DeleteServer(datacenterId, serverId)
	if resp.StatusCode > 299 {
		return errors.New(string(resp.Body))
	}

	err := d.waitTillProvisioned(strings.Join(resp.Headers["Location"], ""))
	if err != nil {
		return err
	}

	resp = profitbricks.DeleteLan(datacenterId, lanId)
	if resp.StatusCode > 299 {
		return errors.New(string(resp.Body))
	}

	err = d.waitTillProvisioned(strings.Join(resp.Headers["Location"], ""))
	if err != nil {
		return err
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

func (d *Driver) waitTillProvisioned(path string) error {
	d.setPB()
	for i := 0; i < waitCount; i++ {
		request := profitbricks.GetRequestStatus(path)
		if request.Metadata.Status == "DONE" {
			return nil
		}
		if request.Metadata.Status == "FAILED" {
			return errors.New(request.Metadata.Message)
		}
		time.Sleep(10 * time.Second)
		i++
	}

	return errors.New("Timeout has expired.")
}

func (d *Driver) getImageId(imageName string) string {
	d.setPB()

	images := profitbricks.ListImages()

	if images.StatusCode == 401 {
		log.Error("Authentication failed")
		return ""
	}

	for _, image := range images.Items {
		imgName := ""
		if image.Properties.Name != "" {
			imgName = image.Properties.Name
		}
		diskType := d.DiskType
		if d.DiskType == "SSD" {
			diskType = "HDD"
		}
		if imgName != "" && strings.Contains(strings.ToLower(imgName), strings.ToLower(imageName)) && image.Properties.ImageType == diskType && image.Properties.Location == d.Location {
			return image.Id
		}
	}
	return ""
}
