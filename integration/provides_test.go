package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testProvides(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image  occam.Image
			name   string
			source string

			pullPolicy       = "never"
			extenderBuildStr = ""
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			if settings.Extensions.UbiNodejsExtension.Online != "" {
				pullPolicy = "always"
				extenderBuildStr = "[extender (build)] "
			}
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("writes a buildplan requiring node and npm", func() {
			var err error

			source, err = occam.Source(filepath.Join("testdata", "needs_node_and_npm_app"))
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
					"      .node-version -> \"20.*\"",
					"      <unknown>     -> \"\"",
				))

				Expect(logs).To(ContainLines(
					"[extender (build)] Enabling module streams:",
					"[extender (build)]     nodejs:20",
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
					MatchRegexp(`    Selected Node Engine version \(using <unknown>\): 20\.\d+\.\d+`),
				))
				Expect(logs).To(ContainLines(
					"  Executing build process",
					MatchRegexp(`    Installing Node Engine 20\.\d+\.\d+`),
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
		})
	})
}
