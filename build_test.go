package nodeengine_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	nodeengine "github.com/paketo-buildpacks/node-engine/v5"
	"github.com/paketo-buildpacks/node-engine/v5/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir        string
		layersDir         string
		cnbDir            string
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		sbomGenerator     *fakes.SBOMGenerator
		buffer            *bytes.Buffer

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "node",
			Metadata: map[string]interface{}{
				"version":        "~10",
				"version-source": "BP_NODE_VERSION",
			},
		}
		entryResolver.MergeLayerTypesCall.Returns.Launch = false
		entryResolver.MergeLayerTypesCall.Returns.Build = false

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{Name: "Node Engine", Version: "10.11.12"}
		// Legacy SBOM
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "node",
				Metadata: paketosbom.BOMMetadata{
					URI:     "node-dependency-uri",
					Version: "~10",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "node-dependency-sha",
					},
				},
			},
		}
		// Syft SBOM
		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		buffer = bytes.NewBuffer(nil)

		build = nodeengine.Build(entryResolver, dependencyManager, sbomGenerator, scribe.NewEmitter(buffer), chronos.DefaultClock)

		buildContext = packit.BuildContext{
			CNBPath: cnbDir,
			Stack:   "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "1.2.3",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"version":        "~10",
							"version-source": "BP_NODE_VERSION",
						},
					},
				},
			},
			Platform:   packit.Platform{Path: "platform"},
			Layers:     packit.Layers{Path: layersDir},
			WorkingDir: workingDir,
		}

	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that installs node", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("node"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "node")))
		Expect(layer.SharedEnv).To(Equal(packit.Environment{
			"NODE_HOME.default":    filepath.Join(layersDir, "node"),
			"NODE_ENV.default":     "production",
			"NODE_VERBOSE.default": "false",
			"NODE_OPTIONS.default": "--use-openssl-ca",
		}))
		Expect(layer.ExecD).To(Equal([]string{
			filepath.Join(cnbDir, "bin", "optimize-memory"),
			filepath.Join(cnbDir, "bin", "inspector"),
		}))

		Expect(layer.Metadata).To(Equal(map[string]interface{}{
			nodeengine.DepKey:    "",
			nodeengine.BuildKey:  false,
			nodeengine.LaunchKey: false,
		}))

		Expect(layer.SBOM.Formats()).To(HaveLen(2))

		cdx := layer.SBOM.Formats()[0]
		spdx := layer.SBOM.Formats()[1]

		Expect(cdx.Extension).To(Equal("cdx.json"))
		content, err := io.ReadAll(cdx.Content)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(MatchJSON(`{
            "$schema": "http://cyclonedx.org/schema/bom-1.3.schema.json",
			"bomFormat": "CycloneDX",
			"metadata": {
				"tools": [
					{
						"name": "",
						"vendor": "anchore"
					}
				]
			},
			"specVersion": "1.3",
			"version": 1
		}`))

		Expect(spdx.Extension).To(Equal("spdx.json"))
		content, err = io.ReadAll(spdx.Content)
		Expect(err).NotTo(HaveOccurred())

		versionPattern := regexp.MustCompile(`"licenseListVersion": "\d+\.\d+"`)
		contentReplaced := versionPattern.ReplaceAllString(string(content), `"licenseListVersion": "x.x"`)

		uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}`)

		contentReplaced = uuidRegex.ReplaceAllString(contentReplaced, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")

		Expect(string(contentReplaced)).To(MatchJSON(`{
			"SPDXID": "SPDXRef-DOCUMENT",
			"creationInfo": {
				"created": "0001-01-01T00:00:00Z",
				"creators": [
					"Organization: Anchore, Inc",
					"Tool: -"
				],
				"licenseListVersion": "x.x"
			},
			"dataLicense": "CC0-1.0",
            "documentNamespace": "https://paketo.io/unknown-source-type/unknown-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			"name": "unknown",
            "packages": [
                {
                  "SPDXID": "SPDXRef-DocumentRoot-Unknown-",
                  "copyrightText": "NOASSERTION",
                  "downloadLocation": "NOASSERTION",
                  "filesAnalyzed": false,
                  "licenseConcluded": "NOASSERTION",
                  "licenseDeclared": "NOASSERTION",
                  "name": "",
                  "supplier": "NOASSERTION"
                }
              ],
			"relationships": [
				{
				    "relatedSpdxElement": "SPDXRef-DocumentRoot-Unknown-",
					"relationshipType": "DESCRIBES",
					"spdxElementId": "SPDXRef-DOCUMENT"
				}
			],
			"spdxVersion": "SPDX-2.2"
		}`))

		Expect(filepath.Join(layersDir, "node")).To(BeADirectory())

		Expect(entryResolver.ResolveCall.Receives.Entries).To(Equal([]packit.BuildpackPlanEntry{
			{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "~10",
					"version-source": "BP_NODE_VERSION",
				},
			},
		}))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Name).To(Equal("node"))
		Expect(entryResolver.MergeLayerTypesCall.Receives.Entries).To(Equal(
			[]packit.BuildpackPlanEntry{
				{
					Name: "node",
					Metadata: map[string]interface{}{
						"version":        "~10",
						"version-source": "BP_NODE_VERSION",
					},
				},
			}))
		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("node"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal("~10"))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine", Version: "10.11.12"}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "node")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))
		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{{Name: "Node Engine", Version: "10.11.12"}}))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dependency).To(Equal(postal.Dependency{Name: "Node Engine", Version: "10.11.12"}))
		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "node")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
		Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using BP_NODE_VERSION): "))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))

		Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/0-optimize-memory"))
		Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
		Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
		Expect(buffer.String()).NotTo(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
		Expect(buffer.String()).NotTo(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
		Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/1-inspector"))
	})

	context("when the os environment contains a directive to optimize memory", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_NODE_OPTIMIZE_MEMORY", "true")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_NODE_OPTIMIZE_MEMORY")).To(Succeed())
		})

		it("tells the environment to optimize memory", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("node"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "node")))
			Expect(layer.LaunchEnv).To(Equal(packit.Environment{
				"OPTIMIZE_MEMORY.default": "true",
			}))

			Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/0-optimize-memory"))
			Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
			Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
			Expect(buffer.String()).To(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
			Expect(buffer.String()).To(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
		})
	})

	context("when the build plan entry includes the build, launch flags", func() {
		var workingDir string

		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "node",
				Metadata: map[string]interface{}{
					"version":        "~10",
					"version-source": "BP_NODE_VERSION",
				},
			}

			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("marks the node layer as build, cached and launch", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("node"))
			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())
			Expect(layer.Metadata).To(HaveKeyWithValue(nodeengine.BuildKey, true))
			Expect(layer.Metadata).To(HaveKeyWithValue(nodeengine.BuildKey, true))
			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "node",
						Metadata: paketosbom.BOMMetadata{
							URI:     "node-dependency-uri",
							Version: "~10",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "node-dependency-sha",
							},
						},
					},
				},
			))
			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "node",
						Metadata: paketosbom.BOMMetadata{
							URI:     "node-dependency-uri",
							Version: "~10",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "node-dependency-sha",
							},
						},
					},
				},
			))
			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{{Name: "Node Engine", Version: "10.11.12"}}))
		})
	})

	context("when there is a dependency cache match", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, "node.toml"), []byte("[metadata]\ndependency-sha = \"some-sha\"\nbuild = false\nlaunch = true\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
				Name:     "Node Engine",
				Checksum: "some-sha",
			}
			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = false
		})

		it("exits build process early", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencyManager.GenerateBillOfMaterialsCall.CallCount).To(Equal(1))
			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
				{
					Name:     "Node Engine",
					Checksum: "some-sha",
				},
			}))
			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "node",
						Metadata: paketosbom.BOMMetadata{
							URI:     "node-dependency-uri",
							Version: "~10",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "node-dependency-sha",
							},
						},
					},
				},
			))
			Expect(result.Build.BOM).To(BeNil())
			Expect(sbomGenerator.GenerateFromDependencyCall.CallCount).To(Equal(0))
			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))
			// Expect(environment.ConfigureCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using BP_NODE_VERSION): "))
			Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			Expect(buffer.String()).ToNot(ContainSubstring("Executing build process"))
		})

		it("the cached layer is NOT used if build requirements do not match", func() {
			entryResolver.MergeLayerTypesCall.Returns.Build = true

			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencyManager.GenerateBillOfMaterialsCall.CallCount).To(Equal(1))
			Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
				{
					Name:     "Node Engine",
					Checksum: "some-sha",
				},
			}))
			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "node",
						Metadata: paketosbom.BOMMetadata{
							URI:     "node-dependency-uri",
							Version: "~10",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "node-dependency-sha",
							},
						},
					},
				},
			))
			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "node",
						Metadata: paketosbom.BOMMetadata{
							URI:     "node-dependency-uri",
							Version: "~10",
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "node-dependency-sha",
							},
						},
					},
				},
			))
			Expect(sbomGenerator.GenerateFromDependencyCall.CallCount).To(Equal(1))
			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(1))
			// Expect(environment.ConfigureCall.CallCount).To(Equal(0))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack 1.2.3"))
			Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
			Expect(buffer.String()).To(ContainSubstring("Selected Node Engine version (using BP_NODE_VERSION): "))
			Expect(buffer.String()).To(ContainSubstring("Installing Node Engine"))
			Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		})

	})

	context("when nodejs has already been provided by an extension", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "",
			}
		})

		it("nodejs layer with environment variables is present", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("node"))
			Expect(layer.Path).To(Equal(filepath.Join(layersDir, "node")))
			Expect(layer.SharedEnv).To(Equal(packit.Environment{
				"NODE_HOME.default":    "",
				"NODE_ENV.default":     "production",
				"NODE_VERBOSE.default": "false",
				"NODE_OPTIONS.default": "--use-openssl-ca",
			}))
			Expect(layer.ExecD).To(Equal([]string{
				filepath.Join(cnbDir, "bin", "optimize-memory"),
				filepath.Join(cnbDir, "bin", "inspector"),
			}))

			Expect(layer.Metadata).To(Equal(map[string]interface{}{
				nodeengine.BuildKey:  true,
				nodeengine.LaunchKey: true,
			}))

			Expect(filepath.Join(layersDir, "node")).To(BeADirectory())

			Expect(buffer.String()).To(ContainSubstring("Resolving Node Engine version"))
			Expect(buffer.String()).To(ContainSubstring("Node no longer requested by plan, satisfied by extension"))

			Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/0-optimize-memory"))
			Expect(buffer.String()).To(ContainSubstring("      Calculates available memory based on container limits at launch time."))
			Expect(buffer.String()).To(ContainSubstring("      Made available in the MEMORY_AVAILABLE environment variable."))
			Expect(buffer.String()).NotTo(ContainSubstring("      Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory."))
			Expect(buffer.String()).NotTo(ContainSubstring("      Limits the total size of all objects on the heap to 75% of the MEMORY_AVAILABLE."))
			Expect(buffer.String()).To(ContainSubstring("    Writing exec.d/1-inspector"))
		})
	})

	context("failure cases", func() {
		context("when a dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when a dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("unsupported SBOM format: 'random-format'"))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})

		context("when BP_DISABLE_SBOM is set incorrectly", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_DISABLE_SBOM", "not-a-bool")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_DISABLE_SBOM")).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to parse BP_DISABLE_SBOM")))
			})
		})

		context("when the layers directory cannot be written to", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the layer directory cannot be removed", func() {
			var layerDir string
			it.Before(func() {
				layerDir = filepath.Join(layersDir, nodeengine.Node)
				Expect(os.MkdirAll(filepath.Join(layerDir, "baller"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(layerDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())
				Expect(os.RemoveAll(layerDir)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

	})
}

func testIsLayerReusable(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		metadata map[string]interface{}
		checksum = "sha256:de15b44738578367cfb250b6551b4c97e0e0e8050fa931a4a9a7262d374d6034/sha256"
		build    = true
		launch   = true
		layer    = packit.Layer{}
		logger   = scribe.NewEmitter(io.Discard)
	)

	it.Before(func() {
		metadata = map[string]interface{}{nodeengine.DepKey: checksum, nodeengine.BuildKey: build, nodeengine.LaunchKey: launch}
		layer.Path = "test"
		layer.Metadata = metadata
	})

	it("returns true if the layer can be reused", func() {
		isReusable := nodeengine.IsLayerReusable(layer, checksum, build, launch, logger)
		Expect(isReusable).To(BeTrue())
	})

	it("returns false if the checksum differs", func() {
		isReusable := nodeengine.IsLayerReusable(layer, "sha256:aaaab44738578367cfb250b6551b4c97e0e0e8050fa931a4a9a7262d3740000/sha256", build, launch, logger)
		Expect(isReusable).To(BeFalse())
	})

	it("returns false if the build requirement changes", func() {
		isReusable := nodeengine.IsLayerReusable(layer, checksum, !build, launch, logger)
		Expect(isReusable).To(BeFalse())
	})

	it("returns false if the launch requirement changes", func() {
		isReusable := nodeengine.IsLayerReusable(layer, checksum, build, !launch, logger)
		Expect(isReusable).To(BeFalse())
	})

	it("returns false if the checksum is missing in metadata", func() {
		delete(metadata, nodeengine.DepKey)
		isReusable := nodeengine.IsLayerReusable(layer, checksum, build, launch, logger)
		Expect(isReusable).To(BeFalse())
	})

	it("returns false if the build is missing in metadata", func() {
		delete(metadata, nodeengine.BuildKey)
		isReusable := nodeengine.IsLayerReusable(layer, checksum, build, launch, logger)
		Expect(isReusable).To(BeFalse())
	})

	it("returns false if the launch is missing in metadata", func() {
		delete(metadata, nodeengine.LaunchKey)
		isReusable := nodeengine.IsLayerReusable(layer, checksum, build, launch, logger)
		Expect(isReusable).To(BeFalse())
	})

}
