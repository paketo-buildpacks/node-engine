package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testSimple(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		pullPolicy       = "never"
		extenderBuildStr = ""
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
			sbomDir   string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			sbomDir, err = os.MkdirTemp("", "sbom")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())

			if settings.Extensions.UbiNodejsExtension.Online != "" {
				pullPolicy = "always"
				extenderBuildStr = "[extender (build)] "
			}
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(os.RemoveAll(sbomDir)).To(Succeed())
		})

		context("simple app", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs correctly", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
					}).
					WithSBOMOutputDir(sbomDir).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				if settings.Extensions.UbiNodejsExtension.Online != "" {

					Expect(logs).To(ContainLines(
						MatchRegexp(`Ubi Node.js Extension \d+\.\d+\.\d+`),
						"  Resolving Node Engine version",
						"    Candidate version sources (in priority order):",
						"      <unknown>     -> \"\"",
					))

					Expect(logs).To(ContainLines(
						"[extender (build)] Enabling module streams:",
						MatchRegexp(`[extender (build)]     nodejs:\d+`),
					))

					Expect(logs).To(ContainLines(
						fmt.Sprintf("[extender (build)] %s 1.2.3", settings.Buildpack.Name),
						"[extender (build)]   Resolving Node Engine version",
						"[extender (build)]   Node no longer requested by plan, satisfied by extension",
					))
				} else {

					Expect(logs).To(ContainLines(
						fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
						"  Resolving Node Engine version",
						"    Candidate version sources (in priority order):",
						"      <unknown> -> \"\"",
					))

					Expect(logs).To(ContainLines(
						MatchRegexp(`    Selected Node Engine version \(using <unknown>\): \d+\.\d+\.\d+`),
					))

					Expect(logs).To(ContainLines(
						"  Executing build process",
						MatchRegexp(`    Installing Node Engine \d+\.\d+\.\d+`),
						MatchRegexp(`      Completed in \d+(\.\d+)?`),
					))
					Expect(logs).To(ContainLines(
						fmt.Sprintf("  Generating SBOM for /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
						MatchRegexp(`      Completed in \d+(\.?\d+)*`),
					))
					Expect(logs).To(ContainLines(
						"  Writing SBOM in the following format(s):",
						"    application/vnd.cyclonedx+json",
						"    application/spdx+json",
						"    application/vnd.syft+json",
					))
				}

				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring build environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"    Writing exec.d/0-optimize-memory",
					extenderBuildStr+"      Calculates available memory based on container limits at launch time.",
					extenderBuildStr+"      Made available in the MEMORY_AVAILABLE environment variable.",
				))

				// Ensure node is installed correctly
				container, err = docker.Container.Run.
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("hello world"))

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("NODE_ENV=production"),
				)

				//ubi does not support SBOM for node at the moment
				if settings.Extensions.UbiNodejsExtension.Online == "" {

					// check that legacy SBOM is included via sbom.legacy.json
					contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", "sbom.legacy.json"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(ContainSubstring(`"name":"Node Engine"`))

					// check that all required SBOM files are present
					Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.cdx.json")).To(BeARegularFile())
					Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.spdx.json")).To(BeARegularFile())
					Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.syft.json")).To(BeARegularFile())

					// check an SBOM file to make sure it has an entry for node
					contents, err = os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.cdx.json"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(ContainSubstring(`"name": "Node Engine"`))

				}
			})
		})

		context("NODE_ENV, NODE_VERBOSE are set by user", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("uses user-set value in build and buildpack-set value in launch phase", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithEnv(map[string]string{"NODE_ENV": "development", "NODE_VERBOSE": "true"}).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring build environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))

				container, err = docker.Container.Run.
					WithCommand("echo ENV=$NODE_ENV && echo VERBOSE=$NODE_VERBOSE && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("hello world"))

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					And(
						ContainSubstring("ENV=production"),
						ContainSubstring("VERBOSE=false"),
					),
				)
			})
		})

		context("simple app with .node-version", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs correctly", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "node_version_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				if settings.Extensions.UbiNodejsExtension.Online != "" {
					Expect(logs).To(ContainLines(
						MatchRegexp(`Ubi Node.js Extension \d+\.\d+\.\d+`),
						"  Resolving Node Engine version",
						"    Candidate version sources (in priority order):",
						"      .node-version -> \"18.*\"",
						"      <unknown>     -> \"\"",
					))

					Expect(logs).To(ContainLines(
						"[extender (build)] Enabling module streams:",
						"[extender (build)]     nodejs:18",
					))

					Expect(logs).To(ContainLines(
						fmt.Sprintf("[extender (build)] %s 1.2.3", settings.Buildpack.Name),
						"[extender (build)]   Resolving Node Engine version",
						"[extender (build)]   Node no longer requested by plan, satisfied by extension",
					))

				} else {
					Expect(logs).To(ContainLines(
						fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
						"  Resolving Node Engine version",
						"    Candidate version sources (in priority order):",
						"      .node-version -> \"18.*\"",
						"      <unknown>     -> \"\"",
					))

					Expect(logs).To(ContainLines(
						MatchRegexp(`    Selected Node Engine version \(using \.node-version\): 18\.\d+\.\d+`),
					))

					Expect(logs).To(ContainLines(
						"  Executing build process",
						MatchRegexp(`    Installing Node Engine 18\.\d+\.\d+`),
						MatchRegexp(`      Completed in \d+(\.\d+)?`),
					))

					Expect(logs).To(ContainLines(
						fmt.Sprintf("  Generating SBOM for /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
						MatchRegexp(`      Completed in \d+(\.?\d+)*`),
					))
				}

				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring build environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"    Writing exec.d/0-optimize-memory",
					extenderBuildStr+"      Calculates available memory based on container limits at launch time.",
					extenderBuildStr+"      Made available in the MEMORY_AVAILABLE environment variable.",
				))

				container, err = docker.Container.Run.
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("hello world"))

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("NODE_ENV=production"),
				)
			})
		})

		context("simple app with .nvmrc", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs correctly", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "nvmrc_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				if settings.Extensions.UbiNodejsExtension.Online != "" {

					Expect(logs).To(ContainLines(
						fmt.Sprintf("[extender (build)] %s 1.2.3", settings.Buildpack.Name),
						"[extender (build)]   Resolving Node Engine version",
						"[extender (build)]   Node no longer requested by plan, satisfied by extension",
					))

				} else {

					Expect(logs).To(ContainLines(
						fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
						"  Resolving Node Engine version",
						"    Candidate version sources (in priority order):",
						"      .nvmrc    -> \"18.*\"",
						"      <unknown> -> \"\"",
					))
					Expect(logs).To(ContainLines(
						MatchRegexp(`    Selected Node Engine version \(using \.nvmrc\): 18\.\d+\.\d+`),
					))
					Expect(logs).To(ContainLines(
						"  Executing build process",
						MatchRegexp(`    Installing Node Engine 18\.\d+\.\d+`),
						MatchRegexp(`      Completed in \d+(\.\d+)?`),
					))

					Expect(logs).To(ContainLines(
						fmt.Sprintf("  Generating SBOM for /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
						MatchRegexp(`      Completed in \d+(\.?\d+)*`),
					))
				}
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring build environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
				Expect(logs).To(ContainLines(
					extenderBuildStr+"    Writing exec.d/0-optimize-memory",
					extenderBuildStr+"      Calculates available memory based on container limits at launch time.",
					extenderBuildStr+"      Made available in the MEMORY_AVAILABLE environment variable.",
				))

				container, err = docker.Container.Run.
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("hello world"))

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("NODE_ENV=production"),
				)
			})
		})

		context("when the node version specfied in the app is EOL'd", func() {
			it("logs thats the dependency is deprecated", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Deprecated,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				// This doesnt apply in ubi-nodejs-extension as there is no
				// such scenario having a deprecated node version in ubi images
				if settings.Extensions.UbiNodejsExtension.Online == "" {
					Expect(logs).To(ContainLines(
						MatchRegexp(`      Version \d+\.\d+\.\d+ of Node Engine is deprecated.`),
						"      Migrate your application to a supported version of Node Engine.",
					))
				}

			})
		})

		context("BP_DISABLE_SBOM is set to true", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("skips SBOM generation", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithSBOMOutputDir(sbomDir).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL":    "DEBUG",
						"BP_DISABLE_SBOM": "true",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				//Below one we ommit it whe using ubi-nodejs-extension
				//as it does not generate SBOM at the moment.
				if settings.Extensions.UbiNodejsExtension.Online == "" {
					Expect(logs).To(ContainLines("    Skipping SBOM generation for Node Engine"))
				}

				// Ensure node is installed correctly
				container, err = docker.Container.Run.
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("hello world"))

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(
					ContainSubstring("NODE_ENV=production"),
				)

				// check that legacy SBOM is NOT included via sbom.legacy.json
				contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", "sbom.legacy.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring("null"))

				// check that SBOM files are not generated
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.cdx.json")).ToNot(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.spdx.json")).ToNot(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "node", "sbom.syft.json")).ToNot(BeARegularFile())
			})
		})
	})
}
