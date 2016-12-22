package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/franela/docker-machine-driver-pwd/pwd"
)

func main() {
	plugin.RegisterDriver(new(pwd.Driver))
}
