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

func testReusingLayerRebuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}
		name         string
		source       string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred())

			firstImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("node"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
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
				fmt.Sprintf("  Generating SBOM for directory /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				"",
				"  Configuring build environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"  Configuring launch environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"    Writing exec.d/0-optimize-memory",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			))

			firstContainer, err = docker.Container.Run.
				WithMemory("128m").
				WithCommand("node server.js").
				WithPublish("8080").
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("node"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Node Engine version \(using <unknown>\): \d+\.\d+\.\d+`),
				"",
				fmt.Sprintf("  Reusing cached layer /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
			))

			secondContainer, err = docker.Container.Run.
				WithMemory("128m").
				WithCommand("node server.js").
				WithPublish("8080").
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort("8080")))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := io.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("hello world"))

			Expect(secondImage.Buildpacks[0].Layers["node"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["node"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred())

			firstImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_NODE_VERSION": "~12"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("node"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      BP_NODE_VERSION -> \"~12\"",
				"      <unknown>       -> \"\"",
				"",
				MatchRegexp(`    Selected Node Engine version \(using BP_NODE_VERSION\): 12\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 12\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				fmt.Sprintf("  Generating SBOM for directory /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				"",
				"  Configuring build environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"  Configuring launch environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"    Writing exec.d/0-optimize-memory",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			))

			firstContainer, err = docker.Container.Run.
				WithMemory("128m").
				WithCommand("node server.js").
				WithPublish("8080").
				Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_NODE_VERSION": "~14"}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(settings.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("node"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      BP_NODE_VERSION -> \"~14\"",
				"      <unknown>       -> \"\"",
				"",
				MatchRegexp(`    Selected Node Engine version \(using BP_NODE_VERSION\): 14\.\d+\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 14\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
				"",
				fmt.Sprintf("  Generating SBOM for directory /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				"",
				"  Configuring build environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"  Configuring launch environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_VERBOSE -> "false"`,
				"",
				"    Writing exec.d/0-optimize-memory",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			))

			secondContainer, err = docker.Container.Run.
				WithMemory("128m").
				WithCommand("node server.js").
				WithPublish("8080").
				Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort("8080")))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := io.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("hello world"))

			Expect(secondImage.Buildpacks[0].Layers["node"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["node"].Metadata["built_at"]))
		})
	})
}
