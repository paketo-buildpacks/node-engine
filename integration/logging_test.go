package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	it("logs useful information for the user", func() {
		app, err := dagger.NewPack(
			filepath.Join("testdata", "simple_app"),
			dagger.RandomImage(),
			dagger.SetBuildpacks(nodeBuildpack),
		).Build()
		Expect(err).ToNot(HaveOccurred())

		buildpackVersion, err := GetGitVersion()
		Expect(err).ToNot(HaveOccurred())

		logs := GetBuildLogs(app.BuildLogs())
		Expect(logs).To(ContainSequence([]interface{}{
			fmt.Sprintf("Node Engine Buildpack %s", buildpackVersion),
			"  Resolving Node Engine version",
			"    Candidate version sources (in priority order):",
			"      buildpack.yml -> \"~10\"",
			"",
			MatchRegexp(`    Selecting Node Engine version \(using buildpack\.yml\): 10\.\d+\.\d+`),
			"",
			"  Executing build process",
			MatchRegexp(`    Installing Node Engine 10\.\d+\.\d+`),
			MatchRegexp(`      Completed in \d+\.\d+`),
			"",
			"    Configuring environment",
			"      NODE_ENV     -> production",
			"      NODE_HOME    -> /layers/org.cloudfoundry.node-engine/node",
			"      NODE_VERBOSE -> false",
			"",
			"      Writing profile.d/0_memory_available.sh",
			"        Calculates available memory based on container limits at launch time.",
			"        Made available in the MEMORY_AVAILABLE environment variable.",
		}))

		app.Destroy()
	})
}
