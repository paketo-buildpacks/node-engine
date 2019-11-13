package main

import (
	"os"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	entryResolver := node.NewPlanEntryResolver(logger)
	dependencyManager := node.NewEngineDependencyManager(cargo.NewTransport(), logger)
	environment := node.NewEnvironment(logger)
	planRefiner := node.NewPlanRefiner()
	cacheHandler := node.NewCacheHandler()

	packit.Build(node.Build(entryResolver, dependencyManager, environment, planRefiner, cacheHandler, logger))
}
