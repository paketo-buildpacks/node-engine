package node

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitNode(t *testing.T) {
	spec.Run(t, "Node", testNode, spec.Report(report.Terminal{}))
}

func testNode(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("NewContributor", func() {
		var stubNodeFixture = filepath.Join("testdata", "stub-node.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{})
			f.AddDependency(Dependency, stubNodeFixture)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})

		it("contributes node to the cache layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true},
			})
			f.AddDependency(Dependency, stubNodeFixture)

			nodeDep, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = nodeDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_HOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_ENV", "production"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_MODULES_CACHE", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_VERBOSE", "false"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_PRODUCTION", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_LOGLEVEL", "error"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_MEMORY", "512"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_CONCURRENCY", "1"))
		})

		it("contributes node to the launch layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(Dependency, stubNodeFixture)

			nodeContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = nodeContributor.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_HOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_HOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_ENV", "production"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_MODULES_CACHE", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_VERBOSE", "false"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_PRODUCTION", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_LOGLEVEL", "error"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_MEMORY", "512"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_CONCURRENCY", "1"))
		})
	})
}
