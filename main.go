package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/play-with-docker/docker-machine-driver-pwd/pwd"
)

func main() {
	plugin.RegisterDriver(new(pwd.Driver))
}
