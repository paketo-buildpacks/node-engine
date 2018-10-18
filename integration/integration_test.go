package integration

import (
	"github.com/sclevine/spec/report"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/dagger"
	"github.com/cloudfoundry/nodejs-cnb/internal/build"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func TestIntegrations(t *testing.T){
	spec.Run(t, "object", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S){
	var (
		rootDir string
		dagg    *dagger.Dagger
	)
	g := NewGomegaWithT(t)

	it.Before(func() {
		var err error

		rootDir, err = dagger.FindRoot()
		g.Expect(err).ToNot(HaveOccurred())

		dagg, err = dagger.NewDagger(rootDir)
		g.Expect(err).ToNot(HaveOccurred())
	})

	it.After(func() {
		dagg.Destroy()
	})

	when("build", func() {
		var group dagger.Group
		it.Before(func() {
			group = dagger.Group{
				Buildpacks: []libbuildpack.BuildpackInfo{
					{
						ID:      "org.cloudfoundry.buildpacks.nodejs",
						Version: "0.0.1",
					},
				},
			}

		})

		when("when the build plan says to add node to the cache", func() {
			var (
				buildResult *dagger.BuildResult
				err         error
			)

			it.Before(func() {
				plan := libbuildpack.BuildPlan{
					build.NodeDependency: libbuildpack.BuildPlanDependency{
						Version: "~10",
						Metadata: libbuildpack.BuildPlanDependencyMetadata{
							"build": true,
						},
					},
				}

				buildResult, err = dagg.Build(filepath.Join(rootDir, "fixtures", "simple_app"), group, plan)
				g.Expect(err).ToNot(HaveOccurred())
			})

			it("installs node in the cache layer", func() {
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "lib")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "include")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "share")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin", "node")).To(BeAnExistingFile())
				g.Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin", "npm")).To(BeAnExistingFile())
			})

			it("sets the nodejs environment variables", func() {
				env, err := buildResult.GetCacheLayerEnv("node")
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(env["NODE_HOME"]).To(Equal("/cache/org.cloudfoundry.buildpacks.nodejs/node"))
				g.Expect(env["NODE_ENV"]).To(Equal("production"))
				g.Expect(env["NODE_MODULES_CACHE"]).To(Equal("true"))
				g.Expect(env["NODE_VERBOSE"]).To(Equal("false"))

				g.Expect(env["NPM_CONFIG_PRODUCTION"]).To(Equal("true"))
				g.Expect(env["NPM_CONFIG_LOGLEVEL"]).To(Equal("error"))

				g.Expect(env["WEB_MEMORY"]).To(Equal("512"))
				g.Expect(env["WEB_CONCURRENCY"]).To(Equal("1"))
			})
		})

		when("when the build plan says to add node to launch", func() {

			var (
				buildResult *dagger.BuildResult
				err         error
			)

			it.Before(func() {
				plan := libbuildpack.BuildPlan{
					build.NodeDependency: libbuildpack.BuildPlanDependency{
						Version: "~10",
						Metadata: libbuildpack.BuildPlanDependencyMetadata{
							"launch": true,
						},
					},
				}

				buildResult, err = dagg.Build(filepath.Join(rootDir, "fixtures", "simple_app"), group, plan)
				g.Expect(err).ToNot(HaveOccurred())
			})

			it("installs node in the launch layer", func() {
				metadata, exists, err := buildResult.GetLayerMetadata("node")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(exists).To(BeTrue())
				g.Expect(metadata.Version).To(MatchRegexp("10.*.*"))

				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "lib")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "include")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "share")).To(BeADirectory())
				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin", "node")).To(BeAnExistingFile())
				g.Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin", "npm")).To(BeAnExistingFile())
			})

			it("sets the nodejs environment variables", func() {
				env, err := buildResult.GetLaunchLayerEnv("node")
				g.Expect(err).ToNot(HaveOccurred())

				g.Expect(env["NODE_HOME"]).To(Equal("/workspace/org.cloudfoundry.buildpacks.nodejs/node"))
				g.Expect(env["NODE_ENV"]).To(Equal("production"))
				g.Expect(env["NODE_MODULES_CACHE"]).To(Equal("true"))
				g.Expect(env["NODE_VERBOSE"]).To(Equal("false"))

				g.Expect(env["NPM_CONFIG_PRODUCTION"]).To(Equal("true"))
				g.Expect(env["NPM_CONFIG_LOGLEVEL"]).To(Equal("error"))

				g.Expect(env["WEB_MEMORY"]).To(Equal("512"))
				g.Expect(env["WEB_CONCURRENCY"]).To(Equal("1"))
			})
		})
	})
}
