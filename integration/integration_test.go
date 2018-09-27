package integration

import (
	"path/filepath"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nodejs buildpack", func() {
	var (
		rootDir string
		dagg    *dagger.Dagger
	)

	BeforeEach(func() {
		var err error

		rootDir, err = dagger.FindRoot()
		Expect(err).ToNot(HaveOccurred())

		dagg, err = dagger.NewDagger(rootDir)
		Expect(err).ToNot(HaveOccurred())

		err = dagg.BundleBuildpack()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dagg.Destroy()
	})

	It("should run detect", func() {
		detectResult, err := dagg.Detect(filepath.Join(rootDir, "fixtures", "simple_app"))
		Expect(err).ToNot(HaveOccurred())

		Expect(len(detectResult.Group.Buildpacks)).To(Equal(1))
		Expect(detectResult.Group.Buildpacks[0].Id).To(Equal("org.cloudfoundry.buildpacks.nodejs"))
		Expect(detectResult.Group.Buildpacks[0].Version).To(Equal("1.6.32"))

		Expect(len(detectResult.BuildPlan)).To(Equal(1))
		Expect(detectResult.BuildPlan).To(HaveKey("node"))
		Expect(detectResult.BuildPlan["node"].Version).To(Equal("~10"))
	})

	It("should run build", func() {
		launchResult, err := dagg.Build(filepath.Join(rootDir, "fixtures", "simple_app"))
		Expect(err).ToNot(HaveOccurred())

		Expect(len(launchResult.LaunchMetadata.Processes)).To(Equal(1))
		Expect(launchResult.LaunchMetadata.Processes[0].Type).To(Equal("web"))
		Expect(launchResult.LaunchMetadata.Processes[0].Command).To(Equal("npm start"))

		nodeLayer := launchResult.Layer
		Expect(nodeLayer.Metadata.Version).To(MatchRegexp("10.*.*"))
		Expect(filepath.Join(nodeLayer.Root, "node", "bin")).To(BeADirectory())
		Expect(filepath.Join(nodeLayer.Root, "node", "lib")).To(BeADirectory())
		Expect(filepath.Join(nodeLayer.Root, "node", "include")).To(BeADirectory())
		Expect(filepath.Join(nodeLayer.Root, "node", "share")).To(BeADirectory())
		Expect(filepath.Join(nodeLayer.Root, "node", "bin", "node")).To(BeAnExistingFile())
		Expect(filepath.Join(nodeLayer.Root, "node", "bin", "npm")).To(BeAnExistingFile())
	})
})
