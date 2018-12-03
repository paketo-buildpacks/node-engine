package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, _ spec.G, it spec.S) {
	it("always passes", func() {
		f := test.NewDetectFactory(t)

		exitStatus, err := runDetect(f.Detect)
		if err != nil {
			t.Fatal(err)
		}

		if exitStatus != detect.PassStatusCode {
			t.Errorf("os.Exit = %d, expected 0", exitStatus)
		}
	})
}
