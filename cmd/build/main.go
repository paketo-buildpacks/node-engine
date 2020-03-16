package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/node-engine-cnb/node"
	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
	"github.com/cloudfoundry/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	entryResolver := node.NewPlanEntryResolver(logger)
	dependencyManager := postal.NewService(cargo.NewTransport())
	environment := node.NewEnvironment(logger)
	planRefiner := node.NewPlanRefiner()
	cacheHandler := node.NewCacheHandler()
	clock := node.NewClock(time.Now)

	packit.Build(node.Build(entryResolver, dependencyManager, environment, planRefiner, cacheHandler, logger, clock))
}
