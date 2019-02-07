package main

import (
	"github.com/orange-cloudfoundry/terraform-secure-backend/cli"
	"os"
)

var Version string

func main() {
	server := cli.NewApp(Version)
	panic(server.Run(os.Args))
}
