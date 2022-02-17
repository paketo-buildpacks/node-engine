package main

import (
	"log"
	"os"

	"github.com/paketo-buildpacks/node-engine/cmd/optimize-memory/internal"
)

func main() {
	err := internal.Run(internal.LoadEnvironmentMap(os.Environ()), os.NewFile(3, "/dev/fd/3"), "/")
	if err != nil {
		log.Fatal(err)
	}
}
