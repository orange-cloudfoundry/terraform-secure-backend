package main

import (
	"github.com/orange-cloudfoundry/terraform-secure-backend/cli"
	"os"
)

func main(){
	server := cli.NewApp()
	server.Run(os.Args)
}