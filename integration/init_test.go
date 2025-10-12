package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpacks struct {
		NodeEngine struct {
			Online     string
			Offline    string
			Deprecated string
		}
		BuildPlan struct {
			Online string
		}
		Processes struct {
			Online string
		}
		Cpython struct {
			Online  string
			Offline string
		}
	}

	Buildpack struct {
		ID   string
		Name string
	}

	Config struct {
		BuildPlan string `json:"build-plan"`
		Cpython   string `json:"cpython"`
	}
}

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.NodeEngine.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Cpython.Online, err = buildpackStore.Get.
		Execute(settings.Config.Cpython)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.Cpython.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(settings.Config.Cpython)
	Expect(err).ToNot(HaveOccurred())

	tmpBuildpackDir, err := os.MkdirTemp("", "node-engine-outdated-deps")
	Expect(err).NotTo(HaveOccurred())

	Expect(cargo.NewDirectoryDuplicator().Duplicate(root, tmpBuildpackDir)).To(Succeed())

	file, err = os.OpenFile(filepath.Join(tmpBuildpackDir, "buildpack.toml"), os.O_RDWR, 0600)
	Expect(err).NotTo(HaveOccurred())

	var buildpackConfig cargo.Config
	_, err = toml.NewDecoder(file).Decode(&buildpackConfig)
	Expect(err).NotTo(HaveOccurred())

	date := time.Date(2000, time.April, 1, 0, 0, 0, 0, time.UTC) // "2000-04-01T00:00:00Z"
	for i := range buildpackConfig.Metadata.Dependencies {
		buildpackConfig.Metadata.Dependencies[i].DeprecationDate = &date
	}

	_, err = file.Seek(0, 0)
	Expect(err).NotTo(HaveOccurred())

	err = file.Truncate(0)
	Expect(err).NotTo(HaveOccurred())

	err = toml.NewEncoder(file).Encode(buildpackConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	settings.Buildpacks.NodeEngine.Deprecated, err = occam.NewBuildpackStore().Get.
		WithVersion("1.2.3").
		Execute(tmpBuildpackDir)
	Expect(err).NotTo(HaveOccurred())
	Expect(os.RemoveAll(tmpBuildpackDir)).To(Succeed())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Processes.Online = filepath.Join("testdata", "processes_buildpack")

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Offline", testOffline)
	suite("OptimizeMemory", testOptimizeMemory)
	suite("Inspector", testInspector)
	suite("ProjectPath", testProjectPath)
	suite("Provides", testProvides)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)
	suite("Simple", testSimple)
	suite("OpenSSL", testOpenSSL)
	suite.Run(t)
}
