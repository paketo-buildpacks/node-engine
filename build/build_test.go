package build_test

import (
	"path/filepath"

	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libjavabuildpack/test"
	"github.com/cloudfoundry/nodejs-cnb/build"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewNode", func() {
	var stubNodeFixture = filepath.Join("lifecycle", "stub-node.tar.gz")

	It("returns true if a build plan exists", func() {
		f := test.NewBuildFactory(T)
		f.AddBuildPlan(T, build.NodeDependency, libbuildpack.BuildPlanDependency{})
		f.AddDependency(T, build.NodeDependency, stubNodeFixture)

		_, ok, err := build.NewNode(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("returns false if a build plan does not exist", func() {
		f := test.NewBuildFactory(T)

		_, ok, err := build.NewNode(f.Build)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("does not contribute node to the cache or launch layer when build and launch are not set", func() {
		f := test.NewBuildFactory(T)
		f.AddBuildPlan(T, build.NodeDependency, libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{},
		})
		f.AddDependency(T, build.NodeDependency, stubNodeFixture)

		nodeDep, _, err := build.NewNode(f.Build)
		Expect(err).NotTo(HaveOccurred())

		err = nodeDep.Contribute()
		Expect(err).NotTo(HaveOccurred())

		cacheLayerRoot := filepath.Join(f.Build.Cache.Root, build.NodeDependency)
		launchLayerRoot := filepath.Join(f.Build.Launch.Root, build.NodeDependency)
		Expect(filepath.Join(cacheLayerRoot, "stub.txt")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(launchLayerRoot, "stub.txt")).NotTo(BeAnExistingFile())
	})

	It("contributes node to the cache layer when build is true", func() {
		f := test.NewBuildFactory(T)
		f.AddBuildPlan(T, build.NodeDependency, libbuildpack.BuildPlanDependency{
			 Metadata: libbuildpack.BuildPlanDependencyMetadata{
			 	"build": true,
			 },
		})
		f.AddDependency(T, build.NodeDependency, stubNodeFixture)

		nodeDep, _, err := build.NewNode(f.Build)
		Expect(err).NotTo(HaveOccurred())

		err = nodeDep.Contribute()
		Expect(err).NotTo(HaveOccurred())

		layerRoot := filepath.Join(f.Build.Cache.Root, build.NodeDependency)
		Expect(filepath.Join(layerRoot, "stub.txt")).To(BeARegularFile())
	})

	It("contributes node to the launch layer when launch is true", func() {
		f := test.NewBuildFactory(T)
		f.AddBuildPlan(T, build.NodeDependency, libbuildpack.BuildPlanDependency{
			Metadata: libbuildpack.BuildPlanDependencyMetadata{
				"launch": true,
			},
		})
		f.AddDependency(T, build.NodeDependency, stubNodeFixture)

		nodeDep, _, err := build.NewNode(f.Build)
		Expect(err).NotTo(HaveOccurred())

		err = nodeDep.Contribute()
		Expect(err).NotTo(HaveOccurred())

		layerRoot := filepath.Join(f.Build.Launch.Root, build.NodeDependency)
		Expect(filepath.Join(layerRoot, "stub.txt")).To(BeARegularFile())
	})
})
