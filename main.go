package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	pwd "github.com/franela/docker-machine-driver-pwd"
)

func main() {
	plugin.RegisterDriver(new(pwd.Driver))
}
