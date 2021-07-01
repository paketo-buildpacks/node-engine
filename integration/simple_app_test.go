package integration

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/cargo"
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
					"      <unknown> -> \"\"",
					"",
					MatchRegexp(`    Selected Node Engine version \(using <unknown>\): \d+\.\d+\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine \d+\.\d+\.\d+`),
					MatchRegexp(`      Completed in \d+\.\d+`),
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
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
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
					WithPullPolicy("never").
					WithEnv(map[string]string{"NODE_ENV": "development", "NODE_VERBOSE": "true"}).
					WithBuildpacks(
						nodeBuildpack,
						buildPlanBuildpack,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					"  Configuring build environment",
					`    NODE_ENV     -> "development"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "true"`,
					"",
					"  Configuring launch environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(config.Buildpack.ID, "/", "_")),
					`    NODE_VERBOSE -> "false"`,
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

				content, err := ioutil.ReadAll(response.Body)
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
					"      .node-version -> \"12.*\"",
					"      <unknown>     -> \"\"",
					"",
					MatchRegexp(`    Selected Node Engine version \(using \.node-version\): 12\.\d+\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine 12\.\d+\.\d+`),
					MatchRegexp(`      Completed in \d+\.\d+`),
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
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
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
					"      .nvmrc    -> \"12.*\"",
					"      <unknown> -> \"\"",
					"",
					MatchRegexp(`    Selected Node Engine version \(using \.nvmrc\): 12\.\d+\.\d+`),
					"",
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine 12\.\d+\.\d+`),
					MatchRegexp(`      Completed in \d+\.\d+`),
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
					WithCommand("echo NODE_ENV=$NODE_ENV && node server.js").
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())

				response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort("8080")))
				Expect(err).NotTo(HaveOccurred())
				Expect(response.StatusCode).To(Equal(http.StatusOK))

				content, err := ioutil.ReadAll(response.Body)
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
			var (
				logs                       fmt.Stringer
				duplicator                 cargo.DirectoryDuplicator
				deprecatedDepNodeBuildpack string
				tmpBuildpackDir            string
			)

			it.Before(func() {
				var err error
				duplicator = cargo.NewDirectoryDuplicator()
				tmpBuildpackDir, err = ioutil.TempDir("", "node-engine-cnb-outdated-deps")
				Expect(err).NotTo(HaveOccurred())

				Expect(duplicator.Duplicate(root, tmpBuildpackDir)).To(Succeed())

				bpToml := []byte(fmt.Sprintf(`
api = "0.2"

[buildpack]
  id = %q
  name = %q

[metadata]
  include-files = ["bin/build", "bin/detect", "bin/run", "buildpack.toml"]
  pre-package = "./scripts/build.sh"
  [metadata.default-versions]
    node = "10.x"

  [[metadata.dependencies]]
		deprecation_date = 2000-04-01T00:00:00Z
    id = "node"
    name = "Node Engine"
    sha256 = "ad0376cbe4dfc3d6092d0ea9fdc4fd3fcb44c477bd4a2c800ccd48eee95e994d"
    source = "https://nodejs.org/dist/v10.18.1/node-v10.18.1.tar.gz"
    source_sha256 = "80a61ffbe6d156458ed54120eb0e9fff7b626502e0986e861d91b365f7e876db"
    stacks = ["some.stack"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node-10.18.1-linux-x64-some-stack-ad0376cb.tgz"
    version = "10.18.1"

  [[metadata.dependencies]]
		deprecation_date = 2000-04-01T00:00:00Z
    id = "node"
    name = "Node Engine"
    sha256 = "528414d1987c8ff9d74f6b5baef604632a2d1d1fbce4a33c7302debcbfa53e1b"
    source = "https://nodejs.org/dist/v10.18.1/node-v10.18.1-linux-x64.tar.gz"
    source_sha256 = "812fe7d421894b792027d19c78c919faad3bf32d8bc16bde67f5c7eea2469eac"
    stacks = ["io.buildpacks.stacks.bionic"]
    uri = "https://buildpacks.cloudfoundry.org/dependencies/node/node-10.18.1-bionic-528414d1.tgz"
    version = "10.18.1"

[[stacks]]
  id = "some.stack"

[[stacks]]
  id = "io.buildpacks.stacks.bionic"
`, config.Buildpack.ID, config.Buildpack.Name))

				err = ioutil.WriteFile(filepath.Join(tmpBuildpackDir, "buildpack.toml"), bpToml, os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				deprecatedDepNodeBuildpack, err = occam.NewBuildpackStore().Get.WithVersion(version).Execute(tmpBuildpackDir)
				Expect(err).NotTo(HaveOccurred())
			})

			it.After(func() {
				os.RemoveAll(tmpBuildpackDir)
			})

			it("logs thats the dependency is deprecated", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).NotTo(HaveOccurred())

				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						deprecatedDepNodeBuildpack,
						buildPlanBuildpack,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs.String()).To(ContainSubstring("Version 10.18.1 of Node Engine is deprecated."))
				Expect(logs.String()).To(ContainSubstring("Migrate your application to a supported version of Node Engine."))
			})
		})
	})
}
