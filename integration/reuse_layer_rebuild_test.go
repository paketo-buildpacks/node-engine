package integration

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReusingLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			app, err := dagger.NewPack(
				filepath.Join("testdata", "simple_app"),
				dagger.RandomImage(),
				dagger.SetBuildpacks(nodeBuildpack),
				dagger.SetVerbose(),
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
				MatchRegexp(`    Selected Node Engine version \(using buildpack\.yml\): 10\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 10\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"  Configuring environment",
				"    NODE_ENV     -> production",
				"    NODE_HOME    -> /layers/org.cloudfoundry.node-engine/node",
				"    NODE_VERBOSE -> false",
				"",
				"    Writing profile.d/0_memory_available.sh",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			}))

			app, err = dagger.NewPack(
				filepath.Join("testdata", "simple_app"),
				dagger.SetImage(app.ImageName),
				dagger.SetBuildpacks(nodeBuildpack),
			).Build()
			Expect(err).ToNot(HaveOccurred())

			logs = GetBuildLogs(app.BuildLogs())
			Expect(logs).To(ContainSequence([]interface{}{
				fmt.Sprintf("Node Engine Buildpack %s", buildpackVersion),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"~10\"",
				"",
				MatchRegexp(`    Selected Node Engine version \(using buildpack\.yml\): 10\.\d+\.\d+`),
				"",
				"  Reusing cached layer /layers/org.cloudfoundry.node-engine/node",
			}))

			Expect(app.StartWithCommand("node server.js")).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("hello world"))

			app.Destroy()
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
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
				MatchRegexp(`    Selected Node Engine version \(using buildpack\.yml\): 10\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 10\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"  Configuring environment",
				"    NODE_ENV     -> production",
				"    NODE_HOME    -> /layers/org.cloudfoundry.node-engine/node",
				"    NODE_VERBOSE -> false",
				"",
				"    Writing profile.d/0_memory_available.sh",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			}))

			app, err = dagger.NewPack(
				filepath.Join("testdata", "different_version_simple_app"),
				dagger.SetImage(app.ImageName),
				dagger.SetBuildpacks(nodeBuildpack),
			).Build()
			Expect(err).ToNot(HaveOccurred())

			logs = GetBuildLogs(app.BuildLogs())
			Expect(logs).To(ContainSequence([]interface{}{
				fmt.Sprintf("Node Engine Buildpack %s", buildpackVersion),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"~12\"",
				"",
				MatchRegexp(`    Selected Node Engine version \(using buildpack\.yml\): 12\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 12\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				"  Configuring environment",
				"    NODE_ENV     -> production",
				"    NODE_HOME    -> /layers/org.cloudfoundry.node-engine/node",
				"    NODE_VERBOSE -> false",
				"",
				"    Writing profile.d/0_memory_available.sh",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			}))
			Expect(app.StartWithCommand("node server.js")).To(Succeed())

			body, _, err := app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(body).To(ContainSubstring("hello world"))

			app.Destroy()
		})
	})
}
