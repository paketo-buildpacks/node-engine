package main

import (
	"testing"

	. "github.com/onsi/gomega"

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
		code, err := runBuild(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(code).To(Equal(build.SuccessStatusCode))
	})
}
