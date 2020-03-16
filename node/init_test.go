package node_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNode(t *testing.T) {
	suite := spec.New("node", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("CacheHandler", testCacheHandler)
	suite("Clock", testClock)
	suite("Detect", testDetect)
	suite("Environment", testEnvironment)
	suite("NvmrcParser", testNvmrcParser)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite("PlanRefinery", testPlanRefinery)
	suite.Run(t)
}
