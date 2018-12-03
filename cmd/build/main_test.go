package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {
	it("always passes", func() {
		f := test.NewBuildFactory(t)

		exitStatus, err := runBuild(f.Build)
		if err != nil {
			t.Fatal(err)
		}

		if exitStatus != build.SuccessStatusCode {
			t.Errorf("os.Exit = %d, expected 0", exitStatus)
		}
	})
}
