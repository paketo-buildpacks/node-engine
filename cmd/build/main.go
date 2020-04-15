package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
	"github.com/paketo-buildpacks/node-engine/node"
)

func main() {
	logEmitter := node.NewLogEmitter(os.Stdout)
	entryResolver := node.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := node.NewEnvironment(logEmitter)
	planRefinery := node.NewPlanRefinery()
	clock := node.NewClock(time.Now)

	packit.Build(node.Build(entryResolver, dependencyManager, environment, planRefinery, logEmitter, clock))
}
