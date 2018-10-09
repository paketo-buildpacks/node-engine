package integration

import (
	"path/filepath"

	"github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/dagger"
	"github.com/cloudfoundry/nodejs-cnb/internal/build"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nodejs buildpack", func() {
	var (
		rootDir string
		dagg    *dagger.Dagger
	)

	BeforeEach(func() {
		var err error

		rootDir, err = dagger.FindRoot()
		Expect(err).ToNot(HaveOccurred())

		dagg, err = dagger.NewDagger(rootDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dagg.Destroy()
	})

	Context("build", func() {
		var group dagger.Group
		BeforeEach(func() {
			group = dagger.Group{
				Buildpacks: []libbuildpack.BuildpackInfo{
					{
						ID:      "org.cloudfoundry.buildpacks.nodejs",
						Version: "1.6.32",
					},
				},
			}

		})

		Context("when the build plan says to add node to the cache", func() {
			var (
				buildResult *dagger.BuildResult
				err         error
			)

			BeforeEach(func() {
				plan := libbuildpack.BuildPlan{
					build.NodeDependency: libbuildpack.BuildPlanDependency{
						Version: "~10",
						Metadata: libbuildpack.BuildPlanDependencyMetadata{
							"build": true,
						},
					},
				}

				buildResult, err = dagg.Build(filepath.Join(rootDir, "fixtures", "simple_app"), group, plan)
				Expect(err).ToNot(HaveOccurred())
			})

			It("installs node in the cache layer", func() {
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin")).To(BeADirectory())
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "lib")).To(BeADirectory())
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "include")).To(BeADirectory())
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "share")).To(BeADirectory())
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin", "node")).To(BeAnExistingFile())
				Expect(filepath.Join(buildResult.CacheRootDir, "node", "bin", "npm")).To(BeAnExistingFile())
			})

			It("sets the nodejs environment variables", func() {
				env, err := buildResult.GetCacheLayerEnv("node")
				Expect(err).ToNot(HaveOccurred())

				Expect(env["NODE_HOME"]).To(Equal("/cache/org.cloudfoundry.buildpacks.nodejs/node"))
				Expect(env["NODE_ENV"]).To(Equal("production"))
				Expect(env["NODE_MODULES_CACHE"]).To(Equal("true"))
				Expect(env["NODE_VERBOSE"]).To(Equal("false"))

				Expect(env["NPM_CONFIG_PRODUCTION"]).To(Equal("true"))
				Expect(env["NPM_CONFIG_LOGLEVEL"]).To(Equal("error"))

				Expect(env["WEB_MEMORY"]).To(Equal("512"))
				Expect(env["WEB_CONCURRENCY"]).To(Equal("1"))
			})
		})

		Context("when the build plan says to add node to launch", func() {

			var (
				buildResult *dagger.BuildResult
				err         error
			)

			BeforeEach(func() {
				plan := libbuildpack.BuildPlan{
					build.NodeDependency: libbuildpack.BuildPlanDependency{
						Version: "~10",
						Metadata: libbuildpack.BuildPlanDependencyMetadata{
							"launch": true,
						},
					},
				}

				buildResult, err = dagg.Build(filepath.Join(rootDir, "fixtures", "simple_app"), group, plan)
				Expect(err).ToNot(HaveOccurred())
			})

			It("installs node in the launch layer", func() {
				metadata, exists, err := buildResult.GetLayerMetadata("node")
				Expect(err).ToNot(HaveOccurred())
				Expect(exists).To(BeTrue())
				Expect(metadata.Version).To(MatchRegexp("10.*.*"))

				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin")).To(BeADirectory())
				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "lib")).To(BeADirectory())
				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "include")).To(BeADirectory())
				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "share")).To(BeADirectory())
				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin", "node")).To(BeAnExistingFile())
				Expect(filepath.Join(buildResult.LaunchRootDir, "node", "bin", "npm")).To(BeAnExistingFile())
			})

			It("sets the nodejs environment variables", func() {
				env, err := buildResult.GetLaunchLayerEnv("node")
				Expect(err).ToNot(HaveOccurred())

				Expect(env["NODE_HOME"]).To(Equal("/workspace/org.cloudfoundry.buildpacks.nodejs/node"))
				Expect(env["NODE_ENV"]).To(Equal("production"))
				Expect(env["NODE_MODULES_CACHE"]).To(Equal("true"))
				Expect(env["NODE_VERBOSE"]).To(Equal("false"))

				Expect(env["NPM_CONFIG_PRODUCTION"]).To(Equal("true"))
				Expect(env["NPM_CONFIG_LOGLEVEL"]).To(Equal("error"))

				Expect(env["WEB_MEMORY"]).To(Equal("512"))
				Expect(env["WEB_CONCURRENCY"]).To(Equal("1"))
			})
		})
	})
})
