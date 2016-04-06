# ProfitBricks Docker Machine Driver

This is the official Docker Machine driver for ProfitBricks

## Install and Run the ProfitBricks Docker Machine Driver

The ProfitBricks plugin will only work with Docker Machine. Before we continue, you will need to install [Docker Machine](https://docs.docker.com/machine/install-machine/)

Next you will need to run the following commands to install the ProfitBricks Docker Machine driver:

    go get github.com/StackPointCloud/docker-machine-driver-profitbricks
    cd $GOPATH/src/github.com/StackPointCloud/docker-machine-driver-profitbricks
    make install  

## Create a ProfitBricks machine

Before you create a ProfitBricks machine you will need to set these variables:

    export PROFITBRICKS_USER_NAME="profitbricks_user_name"
    export PROFITBRICKS_PASSWORD="profitbricks_password"

Then run this command to create a machine:

    docker-machine create --driver profitbricks test-machine

It will produce results similar to this:

```
Running pre-create checks...
Creating machine...
(test-machine) Datacenter Created
(test-machine) Server Created
(test-machine) Volume Created
(test-machine) Attached a volume  to a server.
(test-machine) LAN Created
(test-machine) NIC created
(test-machine) Updated server's boot image
Waiting for machine to be running, this may take a few minutes...
Detecting operating system of created instance...
Waiting for SSH to be available...
Detecting the provisioner...
Provisioning with ubuntu(systemd)...
Installing Docker...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
Checking connection to Docker...
Docker is up and running!
To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: docker-machine env test-machine

```

To get detailed information about the possible options,  run the command:

```docker-machine create --help --driver profitbricks```

```
Usage: docker-machine create [OPTIONS] [arg...]

Create a machine

Description:
   Run 'docker-machine create --driver name' to include the create flags for that driver in the help text.

Options:
   
   --driver, -d "none"											                                    Driver to create machine with.
   --engine-env [--engine-env option --engine-env option]			                                Specify environment variables to set in the engine
   --engine-insecure-registry [--engine-insecure-registry option --engine-insecure-registry option]	Specify insecure registries to allow with the created engine
   --engine-install-url "https://get.docker.com"							                        Custom URL to use for engine installation [$MACHINE_DOCKER_INSTALL_URL]
   --engine-label [--engine-label option --engine-label option]						                Specify labels for the created engine
   --engine-opt [--engine-opt option --engine-opt option]						                    Specify arbitrary flags to include with the created engine in the form flag=value
   --engine-registry-mirror [--engine-registry-mirror option --engine-registry-mirror option]		Specify registry mirrors to use
   --engine-storage-driver 										                                    Specify a storage driver to use with the engine
   --profitbricks-cores "4"										                                    profitbricks cores (2, 3, 4, 5, 6, etc.) [$PROFITBRICKS_CORES]
   --profitbricks-disk-size "50"									                                profitbricks disk size (10, 50, 100, 200, 400) [$PROFITBRICKS_DISK_SIZE]
   --profitbricks-image "Ubuntu-15.10-server-2016-03-01"						                    profitbricks image [$PROFITBRICKS_IMAGE]
   --profitbricks-location "us/lasdev"									                            profitbricks location [$PROFITBRICKS_LOCATION]
   --profitbricks-password 										                                    profitbricks password [$PROFITBRICKS_PASSWORD]
   --profitbricks-ram "2048"										                                profitbricks ram (1024, 2048, 3072, 4096, etc.) [$PROFITBRICKS_RAM]
   --profitbricks-url "https://api.profitbricks.com/rest/v2"					                	profitbricks API endpoint [$PROFITBRICKS_ENDPOINT]
   --profitbricks-user-name 										                                profitbricks user name [$PROFITBRICKS_USER_NAME]
   --swarm												                                            Configure Machine with Swarm
   --swarm-addr 											                                        addr to advertise for Swarm (default: detect and use the machine IP)
   --swarm-discovery 											                                    Discovery service to use with Swarm
   --swarm-host "tcp://0.0.0.0:3376"									                            ip/socket to listen on for Swarm master
   --swarm-image "swarm:latest"										                                Specify Docker image to use for Swarm [$MACHINE_SWARM_IMAGE]
   --swarm-master											                                        Configure Machine to be a Swarm master
   --swarm-opt [--swarm-opt option --swarm-opt option]				                    			Define arbitrary flags for swarm
   --swarm-strategy "spread"										                                Define a default scheduling strategy for Swarm
   --tls-san [--tls-san option --tls-san option]					                        		Support extra SANs for TLS certs
   
```

To list the machines you have created, use the command:

    docker-machine ls

It will return information about your machines, similar to this:

```
NAME           ACTIVE   DRIVER         STATE     URL                         SWARM   DOCKER    ERRORS
default        -        virtualbox     Running   tcp://192.168.99.100:2376           v1.10.2   
test-machine   -        profitbricks   Running   tcp://162.254.26.156:2376           v1.10.3   

```

# Create a Swarm of ProfitBricks Machines 

Before you create a swarm of ProfitBricks machines, run this command:

    docker run --rm swarm create

Then use the output to create the swarm and set a swarm master:

    docker-machine create -d profitbricks --swarm --swarm-master —-swarm-discovery token://f3a75db19a03589ac28550834457bfc3 swarm-master-test

To create a swarm child, use the command:

```docker-machine create -d profitbricks --swarm —-swarm-discovery token://f3a75db19a03589ac28550834457bfc3 swarm-child-test```


## Support

You are welcome to contact us with questions or comments at [ProfitBricks DevOps Central](https://devops.profitbricks.com/). Please report any issues via [GitHub's issue tracker](https://github.com/profitbricks/docker-machine-driver-profitbricks/issues).