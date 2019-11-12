package node_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/node-engine-cnb/node"

	"github.com/cloudfoundry/libcfbuildpack/layers"

	"github.com/sclevine/spec/report"

	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestUnitNode(t *testing.T) {
	spec.Run(t, "Node", testNode, spec.Report(report.Terminal{}))
}

func testNode(t *testing.T, when spec.G, it spec.S) {
	var (
		f               *test.BuildFactory
		stubNodeFixture = filepath.Join("testdata", "stub-node.tar.gz")
	)

	it.Before(func() {
		RegisterTestingT(t)
		f = test.NewBuildFactory(t)
		f.AddDependency(node.Dependency, stubNodeFixture)
	})

	when("node.NewContributor", func() {
		it("returns true if a build plan exists", func() {
			f.AddPlan(buildpackplan.Plan{Name: node.Dependency})

			_, willContribute, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeTrue())
		})

		it("returns false if a build plan does not exist", func() {
			_, willContribute, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())
			Expect(willContribute).To(BeFalse())
		})
	})

	when("Contribute", func() {
		it("writes default env vars, installs the node dependency, writes profile scripts", func() {
			f.AddPlan(buildpackplan.Plan{Name: node.Dependency})

			nodeContributor, _, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(node.Dependency)
			Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_HOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_ENV", "production"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_MODULES_CACHE", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NODE_VERBOSE", "false"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_PRODUCTION", "true"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("NPM_CONFIG_LOGLEVEL", "error"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_MEMORY", "512"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("WEB_CONCURRENCY", "1"))

			memoryAvailableProfile := `if ! which jq > /dev/null; then
	MEMORY_AVAILABLE="$(echo $VCAP_APPLICATION | jq .limits.mem)"
fi

if [[ -z "$MEMORY_AVAILABLE" ]]; then
	memory_in_bytes="$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)"
	MEMORY_AVAILABLE="$(( $memory_in_bytes / ( 1024 * 1024 ) ))"
fi
export MEMORY_AVAILABLE
`
			Expect(layer).To(test.HaveProfile("0_memory_available.sh", memoryAvailableProfile))
		})

		it("uses the default version when a version is not requested", func() {
			f.AddDependencyWithVersion(node.Dependency, "0.9", filepath.Join("testdata", "stub-node-default.tar.gz"))
			f.SetDefaultVersion(node.Dependency, "0.9")
			f.AddPlan(buildpackplan.Plan{Name: node.Dependency})

			nodeContributor, _, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())
			layer := f.Build.Layers.Layer(node.Dependency)
			Expect(layer).To(test.HaveLayerVersion("0.9"))
		})

		it("contributes node to the cache layer when included in the build plan", func() {
			f.AddPlan(buildpackplan.Plan{
				Name: node.Dependency,
				Metadata: buildpackplan.Metadata{
					"build": true,
				},
			})

			nodeContributor, _, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(node.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(true, true, false))
		})

		it("contributes node to the launch layer when included in the build plan", func() {
			f.AddPlan(buildpackplan.Plan{
				Name: node.Dependency,
				Metadata: buildpackplan.Metadata{
					"launch": true,
				},
			})

			nodeContributor, _, err := node.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			Expect(nodeContributor.Contribute()).To(Succeed())

			layer := f.Build.Layers.Layer(node.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
		})

		it("returns an error when unsupported version of node is included in the build plan", func() {
			f.AddPlan(buildpackplan.Plan{
				Name:    node.Dependency,
				Version: "9000.0.0",
				Metadata: buildpackplan.Metadata{
					"launch": true,
				},
			})

			_, shouldContribute, err := node.NewContributor(f.Build)
			Expect(err).To(HaveOccurred())
			Expect(shouldContribute).To(BeFalse())
		})

		when("optimize memory", func() {
			var (
				contributor           node.Contributor
				layer                 layers.Layer
				optimizeMemoryProfile string
				err                   error
			)

			it.Before(func() {
				f.AddPlan(buildpackplan.Plan{
					Name: node.Dependency,
					Metadata: buildpackplan.Metadata{
						"build": true,
					},
				})

				optimizeMemoryProfile = `export NODE_OPTIONS="--max_old_space_size=$(( $MEMORY_AVAILABLE * 75 / 100 ))"`
			})

			it("sets NODE_OPTIONS to use 3/4ths available memory when OPTIMIZE_MEMORY is set", func() {
				contributor, _, err = node.NewContributor(f.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Setenv("OPTIMIZE_MEMORY", "true")).To(Succeed())
				Expect(contributor.Contribute()).To(Succeed())
				layer = f.Build.Layers.Layer(node.Dependency)

				Expect(layer).To(test.HaveProfile("1_optimize_memory.sh", optimizeMemoryProfile))
				Expect(os.Unsetenv("OPTIMIZE_MEMORY")).To(Succeed())
			})

			it("sets NODE_OPTIONS to use 3/4ths available memory when buildpack.yml nodejs.optimize-memory is true", func() {
				yaml := "nodejs:\n  optimize-memory: true"
				test.WriteFile(t, filepath.Join(f.Build.Application.Root, "buildpack.yml"), yaml)
				contributor, _, err = node.NewContributor(f.Build)
				Expect(err).NotTo(HaveOccurred())
				Expect(contributor.Contribute()).To(Succeed())
				layer = f.Build.Layers.Layer(node.Dependency)

				Expect(layer).To(test.HaveProfile("1_optimize_memory.sh", optimizeMemoryProfile))
			})
		})
	})
}
