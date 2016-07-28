package profitbricks

import (
	"io/ioutil"
	"os"
	"testing"
	//"github.com/profitbricks/profitbricks-sdk-go"
	"fmt"
)

const (
	testStoreDir    = ".store-test"
	machineTestName = "docker machine unit tests public 8"
)

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	return d.Data[key].(string)
}

func (d DriverOptionsMock) StringSlice(key string) []string {
	return d.Data[key].([]string)
}

func (d DriverOptionsMock) Int(key string) int {
	return d.Data[key].(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	return d.Data[key].(bool)
}

func cleanup() error {
	return os.RemoveAll(testStoreDir)
}

func getTestStorePath() (string, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return "", err
	}
	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)
	return tmpDir, nil
}

func getDefaultTestDriverFlags() *DriverOptionsMock {
	return &DriverOptionsMock{
		Data: map[string]interface{}{
			"profitbricks-endpoint":  "https://api.profitbricks.com/rest/v2",
			"profitbricks-username":  "user@domain",
			"profitbricks-password":  "password",
			"profitbricks-disk-type": "HDD",
			"profitbricks-disk-size": 5,
			"profitbricks-image":     "Ubuntu-15.10",
			"profitbricks-cores":     1,
			"profitbricks-ram":       1024,
			"profitbricks-location":  "us/lasdev",
			"profitbricks-ssh-key":   `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCoLVLHON4BSK3D8L4H79aFo+0cj7VM2NiRR/K9wrfkK/XiTc7FlEU4Bs8WLZcsIOxbCGWn2zKZmrLaxYlY+/3aJrxDxXYCy8lRUMnqcQ2JCFY6tpZt/DylPhS9L6qYNpJ0F4FlqRsWxsjpF8TDdJi64k2JFJ8TkvX36P2/kqyFfI+N0/axgjhqV3BgNgApvMt9jxWB5gi8LgDpw9b+bHeMS7TrAVDE7bzT86dmfbTugtiME8cIday8YcRb4xAFgRH8XJVOcE3cs390V/dhgCKy1P5+TjQMjKbFIy2LJoxb7bd38kAl1yafZUIhI7F77i7eoRidKV71BpOZsaPEbWUP jasmin@Jasmins-MBP`,
			"swarm-master":           true,
			"swarm-host":             "2",
			"swarm-discovery":        "3",
		},
	}
}

func getTestDriver() (*Driver, error) {
	storePath, err := getTestStorePath()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	d := NewDriver(machineTestName, storePath)

	/*if err != nil {
		return nil, err
	}*/
	d.SetConfigFromFlags(getDefaultTestDriverFlags())
	drv := d.(*Driver)
	return drv, nil
}

func TestCreate(t *testing.T) {
	d, _ := getTestDriver()

	//d.SSHKey, _ = d.createSSHKey("/Users/jasmin/.ssh/id_rsa.pub")
	createerr := d.Create()
	if createerr != nil {
		t.Error(createerr)
	}

	state, err := d.GetState()

	fmt.Println(state)
	if err != nil {
		t.Error(err)
	}
}

func TestGetMachineName(t *testing.T) {
	d, _ := getTestDriver()
	if d.MachineName == "" {
		t.Fatal("Machine name not suplied.")
	}
	fmt.Println(d.GetMachineName())
}

func TestKill(t *testing.T) {
	d, _ := getTestDriver()
	d.Kill()
}

func TestRemove(t *testing.T) {
	d, _ := getTestDriver()
	d.Remove()
}

func TestGetImageName(t *testing.T) {
	d, _ := getTestDriver()
	res := d.getImageId("Debian-8-server1")

	fmt.Println(res == "")
}
