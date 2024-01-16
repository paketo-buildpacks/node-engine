package util_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitUtils(t *testing.T) {
	suite := spec.New("cmd/util", spec.Report(report.Terminal{}))
	suite("EnvironmentMap", testEnvironmentMap)
	suite.Run(t)
}
