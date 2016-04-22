# ProfitBricks Docker Machine Driver

The ProfitBricks Docker Machine Driver is a plugin for Docker Machine which allows you to automate the provisioning of Docker machines on ProfitBricks servers. This is the official Docker Machine driver for ProfitBricks.

### Requirements

*  [Docker Machine](https://docs.docker.com/machine/install-machine/)


## From a Release

The latest version of the `docker-machine-driver-profitbricks` binary is available on the [GithHub Releases](https://github.com/profitbricks/docker-machine-driver-profitbricks/releases) page.
Download the `tar` archive and extract it into a directory residing in your PATH. Select the binary that corresponds to your OS and according to the file name prefix:

* Linux: docker-machine-driver-profitbricks-v1.0.0-linux
* Mac OS X: docker-machine-driver-profitbricks-v1.0.0-darwin-amd64
* Windows: docker-machine-driver-profitbricks-v1.0.0-windows

To extract and install the binary, Linux and Mac users can use the Terminal and the following commands:

```bash
sudo tar -C /usr/local/bin -xvzf docker-machine-driver-profitbricks*.tar.gz
```

If required, modify the permissions to make the plugin executable:

```bash
sudo chmod +x /usr/local/bin/docker-machine-driver-profitbricks
```

Windows users may run the above commands without `sudo` in Docker Quickstart Terminal that is installed with [Docker Toolbox](https://www.docker.com/products/docker-toolbox).


### From Source

Make sure you have installed [Go](http://www.golang.org) and configured [GOPATH](http://golang.org/doc/code.html#GOPATH) properly.

To download the repository and build the driver run the following:

```bash
go get -d -u github.com/profitbricks/docker-machine-driver-profitbricks
cd $GOPATH/src/github.com/profitbricks/docker-machine-driver-profitbricks
make build
```

To use the driver run:

```bash
make install
```

This command will install the driver into `/usr/local/bin`. 

Otherwise, set your PATH environment variable correctly. For example:

```bash
export PATH=$GOPATH/src/github.com/profitbricks/docker-machine-driver-profitbricks/bin:$PATH
```

If you are running Windows, you may also need to install GNU Make, Bash shell and a few other Bash utilities available with [Cygwin](https://www.cygwin.com).

## Usage

You may want to refer to the Docker Machine [official documentation](https://docs.docker.com/machine/) before using the driver.

### Create a ProfitBricks Machine

Before you create a ProfitBricks machine you will need to set these variables:

    export PROFITBRICKS_USERNAME="profitbricks_username"
    export PROFITBRICKS_PASSWORD="profitbricks_password"

Then run this command to create a machine:

    docker-machine create --driver profitbricks test-machine

To get detailed information about the possible options,  run the command:

    docker-machine create --help --driver profitbricks

Available options:

* `--driver`, `-d`: Driver to create machine with.
* `--engine-env [--engine-env option --engine-env option]`: Specify environment variables to set in the engine.
* `--engine-insecure-registry [--engine-insecure-registry option --engine-insecure-registry option]`: Specify insecure registries to allow with the created engine.
* `--engine-install-url` Custom URL to use for engine installation.
* `--engine-label [--engine-label option --engine-label option]`: Specify labels for the created engine.
* `--engine-opt [--engine-opt option --engine-opt option]`: Specify arbitrary flags to include with the created engine in the form flag=value.
* `--engine-registry-mirror [--engine-registry-mirror option --engine-registry-mirror option]`: Specify registry mirrors to use.
* `--engine-storage-driver`: Specify a storage driver to use with the engine
* `--profitbricks-cores`: ProfitBricks cores (2, 3, 4, 5, 6, etc.).
* `--profitbricks-disk-size`: ProfitBricks disk size (10, 50, 100, 200, 400).
* `--profitbricks-disk-type`: ProfitBricks disk type (HDD, SSD).
* `--profitbricks-endpoint`: ProfitBricks API endpoint.
* `--profitbricks-image`: ProfitBricks image.
* `--profitbricks-location`: ProfitBricks location.
* `--profitbricks-password`: ProfitBricks password.
* `--profitbricks-ram`: ProfitBricks RAM (1024, 2048, 3072, 4096, etc.).
* `--profitbricks-username`: ProfitBricks username.
* `--swarm`: Configure Machine with Swarm.
*  `--swarm-addr`: Address to advertise for Swarm.
*  `--swarm-discovery`: Discovery service to use with Swarm.
* `--swarm-host`: IP/socket to listen on for Swarm master.
* `--swarm-image`: Specify Docker image to use for Swarm.
* `--swarm-master`: Configure Machine to be a Swarm master.
* `--swarm-opt [--swarm-opt option --swarm-opt option]`: Define arbitrary flags for Swarm.
* `--swarm-strategy "spread"`: Define a default scheduling strategy for Swarm.
* `--tls-san [--tls-san option --tls-san option]`: Support extra SANs for TLS certs.   

| CLI Option | Default Value | Environment Variable |
|-----------------------------------------------------------------------------------------------------|:--------------------------------------:|:-----------------------------:|
| `--driver`, `-d` | "none" |  |
| `--engine-env  [--engine-env option --engine-env option]` |  |  |
| `--engine-insecure-registry  [--engine-insecure-registry option --engine-insecure-registry option]` |  |  |
| `--engine-install-url,` | https://get.docker.com | [$MACHINE_DOCKER_INSTALL_URL] |
| `--engine-label [--engine-label option --engine-label option]` |  |  |
| `--engine-opt [--engine-opt option --engine-opt option]` |  |  |
| `--engine-registry-mirror  [--engine-registry-mirror option --engine-registry-mirror option]` |  |  |
| `--engine-storage-driver` |  |  |
| `--profitbricks-cores` | 4 | [$PROFITBRICKS_CORES] |
| `--profitbricks-disk-size` | 50 | [$PROFITBRICKS_DISK_SIZE] |
| `--profitbricks-disk-type` | HDD | [$PROFITBRICKS_DISK_TYPE] |
| `--profitbricks-endpoint` | "https://api.profitbricks.com/rest/v2" | [$PROFITBRICKS_ENDPOINT] |
| `--profitbricks-image` | "Ubuntu-15.10-server-2016-03-01" | [$PROFITBRICKS_IMAGE] |
| `--profitbricks-location` | US/LAS | [$PROFITBRICKS_LOCATION] |
| `--profitbricks-password` |  | [$PROFITBRICKS_PASSWORD] |
| `--profitbricks-ram` | 2048 | [$PROFITBRICKS_RAM] |
| `--profitbricks-username` |  | [$PROFITBRICKS_USERNAME] |
| `--swarm` |  |  |
| `--swarm-addr` | Detect and use the machine's IP |  |
| `--swarm-discovery` |  |  |
| `--swarm-host` | tcp://0.0.0.0:3376 |  |
| `--swarm-image` | swarm:latest | [$MACHINE_SWARM_IMAGE] |
| `--swarm-master` |  |  |
| `--swarm-opt [--swarm-opt option --swarm-opt option]` |  |  |
| `--swarm-strategy` | "spread" |  |
| `--tls-san` |  |  |


To list the machines you have created, use the command:

    docker-machine ls

# Create a Swarm of ProfitBricks Machines 

Before you create a swarm of ProfitBricks machines, run this command:

    docker run --rm swarm create

Use the output from this command to create the swarm and set a swarm master:

    docker-machine create -d profitbricks --swarm --swarm-master —-swarm-discovery token://f3a75db19a03589ac28550834457bfc3 swarm-master-test

To create a swarm child, use the command:

```docker-machine create -d profitbricks --swarm —-swarm-discovery token://f3a75db19a03589ac28550834457bfc3 swarm-child-test```


## Support

You are welcome to contact us with questions or comments at [ProfitBricks DevOps Central](https://devops.profitbricks.com/). Please report any issues via [GitHub's issue tracker](https://github.com/profitbricks/docker-machine-driver-profitbricks/issues).
