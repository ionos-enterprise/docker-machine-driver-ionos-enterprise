package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/profitbricks/docker-machine-driver-profitbricks"
)

func main() {
	plugin.RegisterDriver(profitbricks.NewDriver("", ""))
}
