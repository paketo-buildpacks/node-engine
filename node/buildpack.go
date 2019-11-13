package node

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Buildpack struct {
	Info     BuildpackInfo     `toml:"buildpack"`
	Metadata BuildpackMetadata `toml:"metadata"`
}

type BuildpackInfo struct {
	ID      string `toml:"id"`
	Name    string `toml:"name"`
	Version string `toml:"version"`
}

type BuildpackMetadata struct {
	DefaultVersions BuildpackMetadataDefaultVersions `toml:"default-versions"`
	Dependencies    []BuildpackMetadataDependency    `toml:"dependencies"`
}

type BuildpackMetadataDefaultVersions struct {
	Node string `toml:"node"`
}

type BuildpackMetadataDependency struct {
	ID      string                            `toml:"id"`
	Name    string                            `toml:"name"`
	SHA256  string                            `toml:"sha256"`
	Stacks  BuildpackMetadataDependencyStacks `toml:"stacks"`
	URI     string                            `toml:"uri"`
	Version string                            `toml:"version"`
}

type BuildpackMetadataDependencyStacks []string

func ParseBuildpack(path string) (Buildpack, error) {
	var buildpack Buildpack
	_, err := toml.DecodeFile(path, &buildpack)
	if err != nil {
		return Buildpack{}, fmt.Errorf("failed to parse buildpack.toml: %s", err)
	}

	return buildpack, nil
}

func (stacks BuildpackMetadataDependencyStacks) Include(stack string) bool {
	for _, s := range stacks {
		if s == stack {
			return true
		}
	}
	return false
}
