package build_test

import (
	"github.com/sclevine/spec/report"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack/test"
	"github.com/cloudfoundry/nodejs-cnb/internal/build"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitBuild(t *testing.T){
	spec.Run(t, "build", testBuilds, spec.Report(report.Terminal{}))
}

func testBuilds(t *testing.T, when spec.G, it spec.S){
	when("NewNode", func() {
		g:= NewGomegaWithT(t)
		var stubNodeFixture = filepath.Join("lifecycle", "stub-node.tar.gz")

		it("returns true if a build plan exists", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, build.NodeDependency, libbuildpack.BuildPlanDependency{})
			f.AddDependency(t, build.NodeDependency, stubNodeFixture)

			_, ok, err := build.NewNode(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			f := test.NewBuildFactory(t)

			_, ok, err := build.NewNode(f.Build)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ok).To(BeFalse())
		})

		it("does not contribute node to the cache or launch layer when build and launch are not set", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, build.NodeDependency, libbuildpack.BuildPlanDependency{
				Metadata: libbuildpack.BuildPlanDependencyMetadata{},
			})
			f.AddDependency(t, build.NodeDependency, stubNodeFixture)

			nodeDep, _, err := build.NewNode(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			err = nodeDep.Contribute()
			g.Expect(err).NotTo(HaveOccurred())

			cacheLayerRoot := filepath.Join(f.Build.Cache.Root, build.NodeDependency)
			launchLayerRoot := filepath.Join(f.Build.Launch.Root, build.NodeDependency)
			g.Expect(filepath.Join(cacheLayerRoot, "stub.txt")).NotTo(BeAnExistingFile())
			g.Expect(filepath.Join(launchLayerRoot, "stub.txt")).NotTo(BeAnExistingFile())
		})

		it("contributes node to the cache layer when build is true", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, build.NodeDependency, libbuildpack.BuildPlanDependency{
				Metadata: libbuildpack.BuildPlanDependencyMetadata{
					"build": true,
				},
			})
			f.AddDependency(t, build.NodeDependency, stubNodeFixture)

			nodeDep, _, err := build.NewNode(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			err = nodeDep.Contribute()
			g.Expect(err).NotTo(HaveOccurred())

			layerRoot := filepath.Join(f.Build.Cache.Root, build.NodeDependency)
			g.Expect(filepath.Join(layerRoot, "stub.txt")).To(BeARegularFile())
		})

		it("contributes node to the launch layer when launch is true", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(t, build.NodeDependency, libbuildpack.BuildPlanDependency{
				Metadata: libbuildpack.BuildPlanDependencyMetadata{
					"launch": true,
				},
			})
			f.AddDependency(t, build.NodeDependency, stubNodeFixture)

			nodeDep, _, err := build.NewNode(f.Build)
			g.Expect(err).NotTo(HaveOccurred())

			err = nodeDep.Contribute()
			g.Expect(err).NotTo(HaveOccurred())

			layerRoot := filepath.Join(f.Build.Launch.Root, build.NodeDependency)
			g.Expect(filepath.Join(layerRoot, "stub.txt")).To(BeARegularFile())
		})
	})
}
