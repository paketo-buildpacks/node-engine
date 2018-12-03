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
	RegisterTestingT(t)
	spec.Run(t, "Node", testNode, spec.Report(report.Terminal{}))
}

func testNode(t *testing.T, when spec.G, it spec.S) {
	when("NewContributor", func() {
		var stubNodeFixture = filepath.Join("stub-node.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, Dependency, buildplan.Dependency{})
			f.AddDependency(t, Dependency, stubNodeFixture)

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
			f.AddBuildPlan(t, Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": false},
			})
			f.AddDependency(t, Dependency, stubNodeFixture)

			nodeDep, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = nodeDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)
			test.BeLayerLike(t, layer, true, true, false)
			test.BeFileLike(t, filepath.Join(layer.Root, "stub.txt"), 0644, "This is a stub file\n")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_HOME", layer.Root)
			test.BeOverrideSharedEnvLike(t, layer, "NODE_ENV", "production")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_MODULES_CACHE", "true")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_VERBOSE", "false")
			test.BeOverrideSharedEnvLike(t, layer, "NPM_CONFIG_PRODUCTION", "true")
			test.BeOverrideSharedEnvLike(t, layer, "NPM_CONFIG_LOGLEVEL", "error")
			test.BeOverrideSharedEnvLike(t, layer, "WEB_MEMORY", "512")
			test.BeOverrideSharedEnvLike(t, layer, "WEB_CONCURRENCY", "1")
		})

		it("contributes node to the launch layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(t, Dependency, stubNodeFixture)

			nodeContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = nodeContributor.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)
			test.BeLayerLike(t, layer, false, true, true)
			test.BeFileLike(t, filepath.Join(layer.Root, "stub.txt"), 0644, "This is a stub file\n")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_HOME", layer.Root)
			test.BeOverrideSharedEnvLike(t, layer, "NODE_ENV", "production")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_MODULES_CACHE", "true")
			test.BeOverrideSharedEnvLike(t, layer, "NODE_VERBOSE", "false")
			test.BeOverrideSharedEnvLike(t, layer, "NPM_CONFIG_PRODUCTION", "true")
			test.BeOverrideSharedEnvLike(t, layer, "NPM_CONFIG_LOGLEVEL", "error")
			test.BeOverrideSharedEnvLike(t, layer, "WEB_MEMORY", "512")
			test.BeOverrideSharedEnvLike(t, layer, "WEB_CONCURRENCY", "1")
		})
	})
}
