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

func testOpenSSL(t *testing.T, context spec.G, it spec.S) {
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
			sbomDir   string

			pullPolicy       = "never"
			extenderBuildStr = ""
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			sbomDir, err = os.MkdirTemp("", "sbom")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred())
			if settings.Extensions.UbiNodejsExtension.Online != "" {
				pullPolicy = "always"
				extenderBuildStr = "[extender (build)] "
			}
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
			Expect(os.RemoveAll(sbomDir)).To(Succeed())
		})

		context("when running Node 18", func() {
			it("uses the OpenSSL CA store to verify certificates", func() {
				var (
					logs fmt.Stringer
					err  error
				)

				image, logs, err = pack.WithNoColor().Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithPullPolicy(pullPolicy).
					WithEnv(map[string]string{
						"BP_NODE_VERSION": "18.*.*",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithCommand("node server.js").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve("hello world"))
				Expect(container).To(Serve(ContainSubstring("v18.")).WithEndpoint("/version"))
				Expect(container).To(Serve(ContainSubstring("301 Moved")).WithEndpoint("/test-openssl-ca"))

				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))

			})
		})

		context("when running Node 18", func() {
			it("uses the OpenSSL CA store to verify certificates", func() {
				var (
					logs fmt.Stringer
					err  error
				)

				image, logs, err = pack.WithNoColor().Build.
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithPullPolicy(pullPolicy).
					WithEnv(map[string]string{
						"BP_NODE_VERSION": "18.*.*",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithCommand("node server.js").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve("hello world"))
				Expect(container).To(Serve(ContainSubstring("v18.")).WithEndpoint("/version"))
				Expect(container).To(Serve(ContainSubstring("301 Moved")).WithEndpoint("/test-openssl-ca"))

				Expect(logs).To(ContainLines(
					extenderBuildStr+"  Configuring launch environment",
					extenderBuildStr+`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`%s    NODE_HOME    -> "/layers/%s/node"`, extenderBuildStr, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					extenderBuildStr+`    NODE_OPTIONS -> "--use-openssl-ca"`,
					extenderBuildStr+`    NODE_VERBOSE -> "false"`,
				))
			})
		})

		context("when running Node 20", func() {
			it("uses the OpenSSL CA store to verify certificates", func() {
				var (
					logs fmt.Stringer
					err  error
				)

				image, logs, err = pack.WithNoColor().Build.
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithPullPolicy("never").
					WithEnv(map[string]string{
						"BP_NODE_VERSION": "20.*.*",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithPublish("8080").
					WithCommand("node server.js").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve("hello world"))
				Expect(container).To(Serve(ContainSubstring("v20.")).WithEndpoint("/version"))
				Expect(container).To(Serve(ContainSubstring("301 Moved")).WithEndpoint("/test-openssl-ca"))

				Expect(logs).To(ContainLines(
					"  Configuring launch environment",
					`    NODE_ENV     -> "production"`,
					fmt.Sprintf(`    NODE_HOME    -> "/layers/%s/node"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					`    NODE_OPTIONS -> "--use-openssl-ca"`,
					`    NODE_VERBOSE -> "false"`,
				))
			})
		})
	})
}
