package nodeengine_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNode(t *testing.T) {
	suite := spec.New("node", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("NvmrcParser", testNvmrcParser)
	suite("NodeVersionParser", testNodeVersionParser)
	suite.Run(t)
}
