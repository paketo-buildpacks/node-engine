package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	nodeBuildpack        string
	offlineNodeBuildpack string
	buildPlanBuildpack   string
	root                 string
	version              string

	config struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}

	integrationjson struct {
		BuildPlan string `json:"build-plan"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	var err error
	root, err = filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = toml.DecodeReader(file, &config)
	Expect(err).NotTo(HaveOccurred())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&integrationjson)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	version, err = GetGitVersion()
	Expect(err).NotTo(HaveOccurred())

	nodeBuildpack, err = buildpackStore.Get.
		WithVersion(version).
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	offlineNodeBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion(version).
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	buildPlanBuildpack, err = buildpackStore.Get.
		Execute(integrationjson.BuildPlan)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Simple", testSimple)
	suite("Buildpack.yml", testBuildpackYML)
	suite("Offline", testOffline)
	suite("OptimizeMemory", testOptimizeMemory)
	suite("Provides", testProvides)
	suite("ReusingLayerRebuild", testReusingLayerRebuild)
	suite.Run(t)
}

func GetGitVersion() (string, error) {
	gitExec := pexec.NewExecutable("git")
	revListOut := bytes.NewBuffer(nil)

	err := gitExec.Execute(pexec.Execution{
		Args:   []string{"rev-list", "--tags", "--max-count=1"},
		Stdout: revListOut,
	})

	if revListOut.String() == "" {
		return "0.0.0", nil
	}

	if err != nil {
		return "", err
	}

	stdout := bytes.NewBuffer(nil)
	err = gitExec.Execute(pexec.Execution{
		Args:   []string{"describe", "--tags", strings.TrimSpace(revListOut.String())},
		Stdout: stdout,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(stdout.String(), "v")), nil
}
