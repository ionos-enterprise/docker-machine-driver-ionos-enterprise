package main

import (
	"github.com/profitbricks/docker-machine-driver-profitbricks"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(profitbricks.NewDriver("", ""))
}