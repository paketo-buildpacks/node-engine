package nodeengine_test

import (
	"bytes"
	"errors"
	"fmt"
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
                  "SPDXID": "SPDXRef-DocumentRoot-Unknown-unknown",
                  "copyrightText": "NOASSERTION",
                  "downloadLocation": "NOASSERTION",
                  "filesAnalyzed": false,
                  "licenseConcluded": "NOASSERTION",
                  "licenseDeclared": "NOASSERTION",
                  "name": "unknown",
                  "supplier": "NOASSERTION"
                }
              ],
			"relationships": [
				{
				    "relatedSpdxElement": "SPDXRef-DocumentRoot-Unknown-unknown",
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

	context("when the npm bundled in Node.js is upgraded to the pinned npm", func() {
		var (
			layerDir       string
			bundledDir     string
			nodeVersion    string
			bundledVersion string
			pinnedEngines  string
			npmDependency  postal.Dependency
			npmDeliverErr  error
		)

		writeNpm := func(dir, version, engines string) {
			Expect(os.MkdirAll(filepath.Join(dir, "bin"), os.ModePerm)).To(Succeed())
			pkg := fmt.Sprintf(`{"name":"npm","version":%q,"engines":{"node":%q}}`, version, engines)
			Expect(os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(dir, "bin", "npm-cli.js"), []byte(version), 0600)).To(Succeed())
		}

		readNpmVersion := func() string {
			content, err := os.ReadFile(filepath.Join(bundledDir, "package.json"))
			Expect(err).NotTo(HaveOccurred())
			return string(content)
		}

		stagingDirs := func() []string {
			matches, err := filepath.Glob(filepath.Join(layerDir, "lib", "node_modules", ".npm-upgrade-*"))
			Expect(err).NotTo(HaveOccurred())
			return matches
		}

		bomNames := func(entries []packit.BOMEntry) []string {
			var names []string
			for _, entry := range entries {
				names = append(names, entry.Name)
			}
			return names
		}

		it.Before(func() {
			layerDir = filepath.Join(layersDir, nodeengine.Node)
			bundledDir = filepath.Join(layerDir, "lib", "node_modules", "npm")
			nodeVersion = "24.21.0"
			bundledVersion = "11.19.0"
			pinnedEngines = "^20.17.0 || >=22.9.0"
			npmDeliverErr = nil
			npmDependency = postal.Dependency{ID: nodeengine.Npm, Name: "npm", Version: "11.19.1", Checksum: "sha256:npm-sha"}

			dependencyManager.ResolveCall.Stub = func(path, id, version, stack string) (postal.Dependency, error) {
				if id == nodeengine.Npm {
					return npmDependency, nil
				}
				return postal.Dependency{ID: nodeengine.Node, Name: "Node Engine", Version: nodeVersion, Checksum: "sha256:node-sha"}, nil
			}

			// Node.js delivers a bundled npm and corepack; npm delivers the pinned package.
			dependencyManager.DeliverCall.Stub = func(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error {
				if dependency.ID == nodeengine.Npm {
					if npmDeliverErr != nil {
						return npmDeliverErr
					}
					writeNpm(layerPath, dependency.Version, pinnedEngines)
					return nil
				}

				writeNpm(filepath.Join(layerPath, "lib", "node_modules", "npm"), bundledVersion, ">=20")
				Expect(os.MkdirAll(filepath.Join(layerPath, "lib", "node_modules", "corepack"), os.ModePerm)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(layerPath, "bin"), os.ModePerm)).To(Succeed())
				return os.Symlink("../lib/node_modules/npm/bin/npm-cli.js", filepath.Join(layerPath, "bin", "npm"))
			}

			dependencyManager.GenerateBillOfMaterialsCall.Stub = func(dependencies ...postal.Dependency) []packit.BOMEntry {
				var entries []packit.BOMEntry
				for _, dependency := range dependencies {
					entries = append(entries, packit.BOMEntry{Name: dependency.ID})
				}
				return entries
			}

			entryResolver.MergeLayerTypesCall.Returns.Launch = true
			entryResolver.MergeLayerTypesCall.Returns.Build = true
		})

		it("replaces the bundled npm and records it in the BOM and layer metadata", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(dependencyManager.ResolveCall.CallCount).To(Equal(2))
			Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
			Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal(nodeengine.Npm))
			Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
			Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(2))
			Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(npmDependency))
			Expect(dependencyManager.DeliverCall.Receives.LayerPath).To(HavePrefix(filepath.Join(layerDir, "lib", "node_modules", ".npm-upgrade-")))

			Expect(readNpmVersion()).To(ContainSubstring(`"version":"11.19.1"`))
			Expect(filepath.Join(layerDir, "bin", "npm")).To(BeARegularFile())
			Expect(filepath.Join(layerDir, "lib", "node_modules", "corepack")).To(BeADirectory())
			Expect(stagingDirs()).To(BeEmpty())

			// Must stay readable to the app's run user.
			info, err := os.Stat(filepath.Join(layerDir, "lib", "node_modules", "npm"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))

			Expect(result.Layers).To(HaveLen(1))
			Expect(result.Layers[0].Metadata).To(Equal(map[string]interface{}{
				nodeengine.DepKey:         "sha256:node-sha",
				nodeengine.BuildKey:       true,
				nodeengine.LaunchKey:      true,
				nodeengine.NpmDepKey:      "sha256:npm-sha",
				nodeengine.NpmUpgradedKey: true,
			}))

			Expect(bomNames(result.Build.BOM)).To(Equal([]string{"node", "npm"}))
			Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node", "npm"}))

			Expect(buffer.String()).To(ContainSubstring("Checking npm bundled in the Node.js distribution"))
			Expect(buffer.String()).To(ContainSubstring("Installing npm 11.19.1"))
			Expect(buffer.String()).To(ContainSubstring("Upgraded bundled npm 11.19.0 -> 11.19.1"))
		})

		context("when Node.js is not 24.x", func() {
			it.Before(func() {
				nodeVersion = "22.23.2"
				bundledVersion = "10.9.8"
			})

			it("leaves the bundled npm and the layer metadata untouched", func() {
				result, err := build(buildContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(dependencyManager.ResolveCall.CallCount).To(Equal(1))
				Expect(dependencyManager.DeliverCall.CallCount).To(Equal(1))
				Expect(readNpmVersion()).To(ContainSubstring(`"version":"10.9.8"`))

				Expect(result.Layers[0].Metadata).To(Equal(map[string]interface{}{
					nodeengine.DepKey:    "sha256:node-sha",
					nodeengine.BuildKey:  true,
					nodeengine.LaunchKey: true,
				}))
				Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node"}))
				Expect(buffer.String()).NotTo(ContainSubstring("Checking npm bundled in the Node.js distribution"))
			})
		})

		context("when the bundled npm is not older than the pinned npm", func() {
			it.Before(func() {
				bundledVersion = "11.19.1"
			})

			it("keeps the bundled npm", func() {
				result, err := build(buildContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(dependencyManager.DeliverCall.CallCount).To(Equal(1))
				Expect(readNpmVersion()).To(ContainSubstring(`"version":"11.19.1"`))
				Expect(result.Layers[0].Metadata).NotTo(HaveKey(nodeengine.NpmUpgradedKey))
				Expect(result.Layers[0].Metadata).To(HaveKeyWithValue(nodeengine.NpmDepKey, "sha256:npm-sha"))
				Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node"}))
				Expect(buffer.String()).To(ContainSubstring("Keeping bundled npm 11.19.1 (not older than pinned npm 11.19.1)"))
			})
		})

		context("when the bundled npm is a different major version", func() {
			it.Before(func() {
				bundledVersion = "10.9.9"
			})

			it("keeps the bundled npm", func() {
				result, err := build(buildContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(dependencyManager.DeliverCall.CallCount).To(Equal(1))
				Expect(readNpmVersion()).To(ContainSubstring(`"version":"10.9.9"`))
				Expect(result.Layers[0].Metadata).NotTo(HaveKey(nodeengine.NpmUpgradedKey))
				Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node"}))
				Expect(buffer.String()).To(ContainSubstring("Keeping bundled npm 10.9.9 (pinned npm 11.19.1 is a different major version)"))
			})
		})

		context("when the pinned npm does not support the Node.js version", func() {
			it.Before(func() {
				pinnedEngines = ">=26.0.0"
			})

			it("warns and keeps the bundled npm", func() {
				result, err := build(buildContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(dependencyManager.DeliverCall.CallCount).To(Equal(2))
				Expect(readNpmVersion()).To(ContainSubstring(`"version":"11.19.0"`))
				Expect(stagingDirs()).To(BeEmpty())
				Expect(result.Layers[0].Metadata).NotTo(HaveKey(nodeengine.NpmUpgradedKey))
				Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node"}))
				Expect(buffer.String()).To(ContainSubstring(`WARNING: npm 11.19.1 requires Node.js ">=26.0.0", which does not include 24.21.0`))
				Expect(buffer.String()).To(ContainSubstring("Keeping bundled npm 11.19.0"))
			})
		})

		context("when the pinned npm cannot be installed", func() {
			it.Before(func() {
				npmDeliverErr = errors.New("failed to deliver npm")
			})

			it("returns an error and leaves the bundled npm in place", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to deliver npm"))

				Expect(readNpmVersion()).To(ContainSubstring(`"version":"11.19.0"`))
				Expect(filepath.Join(layerDir, "bin", "npm")).To(BeARegularFile())
				Expect(stagingDirs()).To(BeEmpty())
			})
		})

		context("when the pinned npm cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Stub = func(path, id, version, stack string) (postal.Dependency, error) {
					if id == nodeengine.Npm {
						return postal.Dependency{}, errors.New("no npm dependency")
					}
					return postal.Dependency{ID: nodeengine.Node, Name: "Node Engine", Version: nodeVersion, Checksum: "sha256:node-sha"}, nil
				}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve pinned npm for Node.js 24.21.0: no npm dependency"))
			})
		})

		context("when the layer was cached with the upgraded npm", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layersDir, "node.toml"), []byte(`[metadata]
dependency-sha = "sha256:node-sha"
npm-dependency-sha = "sha256:npm-sha"
npm-upgraded = true
build = true
launch = true
`), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			it("reuses the layer and keeps npm in the BOM", func() {
				result, err := build(buildContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))
				Expect(bomNames(result.Build.BOM)).To(Equal([]string{"node", "npm"}))
				Expect(bomNames(result.Launch.BOM)).To(Equal([]string{"node", "npm"}))
				Expect(buffer.String()).To(ContainSubstring("Reusing cached layer"))
			})

			context("when the pinned npm has changed", func() {
				it.Before(func() {
					npmDependency.Checksum = "sha256:new-npm-sha"
				})

				it("rebuilds the layer", func() {
					result, err := build(buildContext)
					Expect(err).NotTo(HaveOccurred())

					Expect(dependencyManager.DeliverCall.CallCount).To(Equal(2))
					Expect(result.Layers[0].Metadata).To(HaveKeyWithValue(nodeengine.NpmDepKey, "sha256:new-npm-sha"))
					Expect(buffer.String()).To(ContainSubstring("Executing build process"))
					Expect(buffer.String()).NotTo(ContainSubstring("Reusing cached layer"))
				})
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

func testIsNpmUpgradeReusable(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layer  packit.Layer
		logger = scribe.NewEmitter(io.Discard)
	)

	it.Before(func() {
		layer = packit.Layer{
			Path:     "test",
			Metadata: map[string]interface{}{nodeengine.NpmDepKey: "sha256:npm-sha"},
		}
	})

	it("returns true if the pinned npm checksum matches", func() {
		Expect(nodeengine.IsNpmUpgradeReusable(layer, "sha256:npm-sha", logger)).To(BeTrue())
	})

	it("returns false if the pinned npm checksum differs", func() {
		Expect(nodeengine.IsNpmUpgradeReusable(layer, "sha256:other-sha", logger)).To(BeFalse())
	})

	it("returns false if the layer has no npm checksum but one is pinned", func() {
		delete(layer.Metadata, nodeengine.NpmDepKey)
		Expect(nodeengine.IsNpmUpgradeReusable(layer, "sha256:npm-sha", logger)).To(BeFalse())
	})

	it("returns false if the layer has an npm checksum but none applies", func() {
		Expect(nodeengine.IsNpmUpgradeReusable(layer, "", logger)).To(BeFalse())
	})

	it("returns true if neither the layer nor the build uses the npm upgrade", func() {
		delete(layer.Metadata, nodeengine.NpmDepKey)
		Expect(nodeengine.IsNpmUpgradeReusable(layer, "", logger)).To(BeTrue())
	})
}
