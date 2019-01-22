package main

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, _ spec.G, it spec.S) {
	it("always passes", func() {
		f := test.NewDetectFactory(t)
		code, err := runDetect(f.Detect)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(detect.PassStatusCode))
	})
}
