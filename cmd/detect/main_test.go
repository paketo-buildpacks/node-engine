package main

import (
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/nodejs-cnb/node"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("always passes", func() {
		f := test.NewDetectFactory(t)
		runDetectAndExpectBuildplan(f, buildplan.BuildPlan{})
	})

	when("there is an .nvmrc", func () {
		it("should request the node version in the .nvmrc file", func() {
			f := test.NewDetectFactory(t)

			test.WriteFile(t, filepath.Join(f.Detect.Application.Root, ".nvmrc"), "8.15.0")
			buildplan := getStandardBuildplanWithNodeVersion("8.15.0")
			runDetectAndExpectBuildplan(f, buildplan)
		})
	})
}

func runDetectAndExpectBuildplan(factory *test.DetectFactory, buildplan buildplan.BuildPlan) {
	code, err := runDetect(factory.Detect)
	Expect(err).NotTo(HaveOccurred())

	Expect(code).To(Equal(detect.PassStatusCode))

	Expect(factory.Output).To(Equal(buildplan))
}

func getStandardBuildplanWithNodeVersion(version string) buildplan.BuildPlan {
	return buildplan.BuildPlan{
		node.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"launch": true},
		},
	}
}