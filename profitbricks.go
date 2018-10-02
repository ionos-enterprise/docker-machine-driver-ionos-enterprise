package profitbricks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/profitbricks/profitbricks-sdk-go"
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
	UseAlias               bool
	LanId                  string
	client                 *profitbricks.Client
}

const (
	defaultRegion = "us/las"
	defaultSize   = 10
	waitCount     = 1000
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "PROFITBRICKS_ENDPOINT",
			Name:   "profitbricks-endpoint",
			Value:  "https://api.profitbricks.com/cloudapi/v4",
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
			Name:  "profitbricks-datacenter-id",
			Usage: "ProfitBricks Virtual Data Center Id",
		},
		mcnflag.StringFlag{
			Name:  "profitbricks-volume-availability-zone",
			Value: "AUTO",
			Usage: "ProfitBricks Volume Availability Zone (AUTO, ZONE_1, ZONE_2, ZONE_3)",
		},
		mcnflag.StringFlag{
			Name:  "profitbricks-server-availability-zone",
			Value: "AUTO",
			Usage: "ProfitBricks Server Availability Zone (AUTO, ZONE_1, ZONE_2, ZONE_3)",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	driver := &Driver{
		Size:     defaultSize,
		Location: defaultRegion,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
	driver.setPB()

	return driver
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
		d.URL = "https://api.profitbricks.com/cloudapi/v4"
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.Username == "" {
		return fmt.Errorf("Please provide username as paramter --profitbricks-username or as environment variable $PROFITBRICKS_USERNAME")
	}
	if d.DatacenterId != "" {
		dc, err := d.client.GetDatacenter(d.DatacenterId)

		if err != nil {
			return fmt.Errorf("An error occurred while fetching datacenter '%s': %s", d.DatacenterId, dc.Response)
		}
		log.Info("Creating machine under " + dc.Properties.Name + " datacenter.")
	}

	if d.getImageId(d.Image) == "" {
		return fmt.Errorf("The image/alias  %s %s %s", d.Image, d.Location, "does not exist.")
	}

	return nil
}

func (d *Driver) Create() error {
	var err error
	var image string
	var alias string
	if d.SSHKey == "" {
		d.SSHKey, err = d.createSSHKey()
		if err != nil {
			return err
		}
	}
	var result = d.getImageId(d.Image)
	if !d.UseAlias {
		image = result
	} else {
		alias = result
	}

	ipblockreq := profitbricks.IPBlock{
		Properties: profitbricks.IPBlockProperties{
			Size:     1,
			Location: d.Location,
		},
	}

	ipblockresp, err := d.client.ReserveIPBlock(ipblockreq)

	if err != nil {
		return fmt.Errorf("An error occurred while reserving an ipblock: %s", ipblockresp.Response)
	}

	err = d.waitTillProvisioned(ipblockresp.Headers.Get("Location"))
	if err != nil {
		return err
	}

	var dc *profitbricks.Datacenter

	if d.DatacenterId == "" {
		d.DCExists = false
		req := profitbricks.Datacenter{
			Properties: profitbricks.DatacenterProperties{
				Name:     d.MachineName,
				Location: d.Location,
			},
		}

		dc, err = d.client.CreateDatacenter(req)
		if err == nil {
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
		dc, _ = d.client.GetDatacenter(d.DatacenterId)
	}

	lan, _ := d.client.CreateLan(dc.ID, profitbricks.Lan{
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

	d.DatacenterId = dc.ID

	err = d.waitTillProvisioned(lan.Headers.Get("Location"))
	if err != nil {
		return err
	}

	lanId, _ := strconv.Atoi(lan.ID)

	d.LanId = lan.ID
	serverReq := profitbricks.Server{
		Properties: profitbricks.ServerProperties{
			Name:             d.MachineName,
			RAM:              d.Ram,
			Cores:            d.Cores,
			CPUFamily:        d.CpuFamily,
			AvailabilityZone: d.ServerAvailabilityZone,
		},
		Entities: &profitbricks.ServerEntities{
			Volumes: &profitbricks.Volumes{
				Items: []profitbricks.Volume{
					{
						Properties: profitbricks.VolumeProperties{
							Type:             d.DiskType,
							Size:             d.DiskSize,
							Name:             d.MachineName,
							Image:            image,
							ImageAlias:       alias,
							SSHKeys:          []string{d.SSHKey},
							AvailabilityZone: d.VolumeAvailabilityZone,
						},
					},
				},
			},
		},
	}

	dhcp := true
	nic := profitbricks.Nic{
		Properties: &profitbricks.NicProperties{
			Name: d.MachineName,
			Lan:  lanId,
			Ips:  ipblockresp.Properties.IPs,
			Dhcp: &dhcp,
		},
	}

	serverReq.Entities.Nics = &profitbricks.Nics{
		Items: []profitbricks.Nic{
			nic,
		},
	}

	server, err := d.client.CreateServer(dc.ID, serverReq)

	if err == nil {
		log.Info("Server Created")
	} else {
		d.Remove()
		return errors.New("Error while creating a server " + string(server.Response) + "Rolling back...")
	}

	err = d.waitTillProvisioned(server.Headers.Get("Location"))
	if err != nil {
		return err
	}
	d.ServerId = server.ID

	d.IPAddress = ipblockresp.Properties.IPs[0]
	log.Info(d.IPAddress)
	return nil
}

func (d *Driver) Restart() error {
	_, err := d.client.RebootServer(d.DatacenterId, d.ServerId)
	return err
}

func (d *Driver) Remove() error {
	if !d.DCExists {
		servers, _ := d.client.ListServers(d.DatacenterId)
		if len(servers.Items) == 1 {
			resp, err := d.client.DeleteDatacenter(d.DatacenterId)
			if err != nil {
				return err
			}

			err = d.waitTillProvisioned(resp.Get("Location"))
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

	ipblocks, _ := d.client.ListIPBlocks()

	for _, i := range ipblocks.Items {
		for _, v := range i.Properties.IPs {
			if d.IPAddress == v {
				_, err := d.client.ReleaseIPBlock(i.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (d *Driver) removeServer(datacenterId string, serverId string, lanId string) error {
	server, _ := d.client.GetServer(datacenterId, serverId)

	if server.StatusCode > 299 {
		return errors.New(server.Response)
	}

	if server.Entities != nil && server.Entities.Volumes != nil && len(server.Entities.Volumes.Items) > 0 {
		volumeId := server.Entities.Volumes.Items[0].ID
		resp, err := d.client.DeleteVolume(d.DatacenterId, volumeId)
		if err != nil {
			return err
		}
		err = d.waitTillProvisioned(resp.Get("Location"))

		if err != nil {
			return err
		}
	}
	resp, err := d.client.DeleteServer(datacenterId, serverId)
	if err != nil {
		return err
	}

	err = d.waitTillProvisioned(resp.Get("Location"))
	if err != nil {
		return err
	}

	resp, err = d.client.DeleteLan(datacenterId, lanId)
	if err != nil {
		return err
	}

	err = d.waitTillProvisioned(resp.Get("Location"))
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

	if err != nil {
		return err
	}
	if serverstate != state.Running {
		_, err = d.client.StartServer(d.DatacenterId, d.ServerId)
	} else {
		log.Info("Host is already running or starting")
	}
	return err
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

	_, err = d.client.StopServer(d.DatacenterId, d.ServerId)

	return err
}

func (d *Driver) Kill() error {
	_, err := d.client.StopServer(d.DatacenterId, d.ServerId)
	return err
}

func (d *Driver) GetIP() (string, error) {
	server, _ := d.client.GetServer(d.DatacenterId, d.ServerId)

	d.IPAddress = server.Entities.Nics.Items[0].Properties.Ips[0]
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	server, _ := d.client.GetServer(d.DatacenterId, d.ServerId)

	if server.StatusCode > 299 {

		if server.StatusCode == 401 {
			return state.None, fmt.Errorf("Unauthorized. Either user name or password are incorrect.")

		} else {
			return state.None, fmt.Errorf("Error occurred while fetching a server: %s", server.Response)
		}
	}

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
	client := profitbricks.NewClient(d.Username, d.Password)
	client.SetUserAgent(fmt.Sprintf("%s", d.client.GetUserAgent()+"docker-machine-driver-profitbricks/1.3.4"))
	client.SetURL(d.URL)
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
	for i := 0; i < waitCount; i++ {
		request, err := d.client.GetRequestStatus(path)
		if err != nil {
			return err
		}
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
	d.UseAlias = false
	//first look if the provided parameter matches an alias, if a match is found we return the image alias
	location, _ := d.client.GetLocation(d.Location)

	for _, alias := range location.Properties.ImageAliases {
		if alias == imageName {
			d.UseAlias = true
			return imageName
		}
	}

	//if no alias matchs we do extended search and return the image id
	images, _ := d.client.ListImages()

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
			return image.ID
		}
	}
	return ""
}
