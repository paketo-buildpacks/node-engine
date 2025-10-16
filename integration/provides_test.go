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

		it("writes a buildplan requiring node and npm", func() {
			var err error

			source, err = occam.Source(filepath.Join("testdata", "needs_node_and_npm_app"))
			Expect(err).ToNot(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				fmt.Sprintf("%s 1.2.3", settings.Buildpack.Name),
				"  Resolving Node Engine version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected Node Engine version \(using <unknown>\): 22\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing Node Engine 22\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+(\.\d+)?`),
			))
			Expect(logs).To(ContainLines(
				fmt.Sprintf("  Generating SBOM for /layers/%s/node", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
			))
			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_OPTIONS -> "--use-openssl-ca"`,
				`    NODE_VERBOSE -> "false"`,
			))
			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				`    NODE_ENV     -> "production"`,
				fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
				`    NODE_OPTIONS -> "--use-openssl-ca"`,
				`    NODE_VERBOSE -> "false"`,
			))
			Expect(logs).To(ContainLines(
				"    Writing exec.d/0-optimize-memory",
				"      Calculates available memory based on container limits at launch time.",
				"      Made available in the MEMORY_AVAILABLE environment variable.",
			))
		})
	})
}
