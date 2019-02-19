package main

import (
	"github.com/cloudfoundry/cnb-tools/install_tools"
	"github.com/cloudfoundry/cnb-tools/integration"
	"github.com/cloudfoundry/cnb-tools/package"
	"github.com/cloudfoundry/cnb-tools/unit"
	"log"
	"os"
)

func main() {
	args := len(os.Args)
	if args < 2 || os.Args[1] == "" {
		log.Fatal("Please enter a script name")
	}

	scriptName := os.Args[1]

	if err := initCmd(scriptName); err != nil {
		log.Fatal(err)
	}
}

func initCmd(name string) error {
	var err error
	switch name {
	case "unit":
		err = unit_cnb.Run()
	case "integration":
		err = integration_cnb.Run()
	case "package":
		err = package_cnb.Run()
	case "install_tools":
		err = install_tools_cnb.Run()
	default:
		log.Fatal("Please enter a script name")
	}
	if err != nil {
		return err
	}
	return nil
}
