package internal_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitInspector(t *testing.T) {
	suite := spec.New("cmd/inspector/internal", spec.Report(report.Terminal{}))
	suite("Run", testRun)
	suite.Run(t)
}
