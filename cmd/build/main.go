package main

import (
	"os"

	"github.com/paketo-buildpacks/node-engine/node"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	logEmitter := node.NewLogEmitter(os.Stdout)
	entryResolver := node.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := node.NewEnvironment(logEmitter)
	planRefinery := node.NewPlanRefinery()

	packit.Build(node.Build(entryResolver, dependencyManager, environment, planRefinery, logEmitter, chronos.DefaultClock))
}
