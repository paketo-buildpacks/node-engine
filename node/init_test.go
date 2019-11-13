package node_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNode(t *testing.T) {
	suite := spec.New("node", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Buildpack", testBuildpack)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("CacheHandler", testCacheHandler)
	suite("Detect", testDetect)
	suite("EngineDependencyManager", testEngineDependencyManager)
	suite("Environment", testEnvironment)
	suite("NvmrcParser", testNvmrcParser)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite("PlanRefiner", testPlanRefiner)
	suite.Run(t)
}
