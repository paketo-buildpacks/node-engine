package main

import (
	"log"
	"os"

	"github.com/paketo-buildpacks/node-engine/v3/cmd/optimize-memory/internal"
	"github.com/paketo-buildpacks/node-engine/v3/cmd/util"
)

func main() {
	err := internal.Run(util.LoadEnvironmentMap(os.Environ()), os.NewFile(3, "/dev/fd/3"), "/")
	if err != nil {
		log.Fatal(err)
	}
}
