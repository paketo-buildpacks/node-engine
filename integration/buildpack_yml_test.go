package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackYML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("simple app with version and optimize memory set in buildpack.yml", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs correctly but gives a deprecation warning", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "buildpack_yml_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						nodeBuildpack,
						buildPlanBuildpack,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					fmt.Sprintf("%s %s", config.Buildpack.Name, version),
					"  Resolving Node Engine version",
					"    Candidate version sources (in priority order):",
					"      buildpack.yml -> \"~12\"",
					"      <unknown>     -> \"\"",
					"",
					MatchRegexp(`    Selected Node Engine version \(using buildpack\.yml\): 12\.\d+\.\d+`),
					"",
					"    WARNING: Setting the Node version through buildpack.yml will be deprecated soon in Node Engine Buildpack v1.0.0.",
					"    Please specify the version through the $BP_NODE_VERSION environment variable instead. See README.md for more information.",
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine 12\.\d+\.\d+`),
					MatchRegexp(`      Completed in \d+\.\d+`),
					"",
					"    WARNING: Enabling memory optimization through buildpack.yml will be deprecated soon in Node Engine Buildpack v1.0.0.",
					"    Please enable through the $BP_NODE_OPTIMIZE_MEMORY environment variable instead. See README.md for more information.",
					"",
					"  Configuring build environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "false"`,
					"",
					"  Configuring launch environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "false"`,
					"",
					"    Writing profile.d/0_memory_available.sh",
					"      Calculates available memory based on container limits at launch time.",
					"      Made available in the MEMORY_AVAILABLE environment variable.",
					"",
					"    Writing profile.d/1_optimize_memory.sh",
					"      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.",
					"      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE.",
				))

				container, err = docker.Container.Run.
					WithMemory("128m").
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("NodeOptions: --max_old_space_size=96")).OnPort(8080))

			})
		})

		context("app with BP_NODE_VERSION set and a buildpack.yml", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs correctly with BP_NODE_VERSION", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "buildpack_yml_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						nodeBuildpack,
						buildPlanBuildpack,
					).
					WithEnv(map[string]string{"BP_NODE_VERSION": "~14.15"}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					fmt.Sprintf("%s %s", config.Buildpack.Name, version),
					"  Resolving Node Engine version",
					"    Candidate version sources (in priority order):",
					"      BP_NODE_VERSION -> \"~14.15\"",
					"      buildpack.yml   -> \"~12\"",
					"      <unknown>       -> \"\"",
					"",
					MatchRegexp(`    Selected Node Engine version \(using BP_NODE_VERSION\): 14\.15\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine 14\.15\.\d+`),
					MatchRegexp(`      Completed in \d+\.\d+`),
					"",
					"    WARNING: Enabling memory optimization through buildpack.yml will be deprecated soon in Node Engine Buildpack v1.0.0.",
					"    Please enable through the $BP_NODE_OPTIMIZE_MEMORY environment variable instead. See README.md for more information.",
					"",
					"  Configuring build environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "false"`,
					"",
					"  Configuring launch environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "false"`,
					"",
					"    Writing profile.d/0_memory_available.sh",
					"      Calculates available memory based on container limits at launch time.",
					"      Made available in the MEMORY_AVAILABLE environment variable.",
				))

				container, err = docker.Container.Run.
					WithMemory("128m").
					WithCommand("node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("NodeOptions: --max_old_space_size=96")).OnPort(8080))
			})
		})
	})
}
