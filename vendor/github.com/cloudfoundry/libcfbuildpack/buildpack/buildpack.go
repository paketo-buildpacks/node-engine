/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buildpack

import (
	"fmt"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

const (
	cacheRoot            = "dependency-cache"
	DependenciesMetadata = "dependencies"
	DefaultVersions      = "default-versions"
)

// Buildpack is an extension to libbuildpack.Buildpack that adds additional opinionated behaviors.
type Buildpack struct {
	buildpack.Buildpack

	// CacheRoot is the path to the root directory for the buildpack's dependency cache.
	CacheRoot string

	logger logger.Logger
}

// Dependencies returns the collection of dependencies extracted from the generic buildpack metadata.
func (b Buildpack) Dependencies() (Dependencies, error) {
	deps, ok := b.Metadata[DependenciesMetadata].([]map[string]interface{})
	if !ok {
		return Dependencies{}, nil
	}

	var dependencies Dependencies
	for _, dep := range deps {
		d, err := NewDependency(dep)
		if err != nil {
			return Dependencies{}, err
		}

		dependencies = append(dependencies, d)
	}

	b.logger.Debug("Dependencies: %s", dependencies)
	return dependencies, nil
}

func (b Buildpack) DefaultVersion(id string) (string, error) {
	defaults, ok := b.Metadata[DefaultVersions].(map[string]interface{})
	if !ok {
		return "", nil
	}

	version, ok := defaults[id].(string)
	if !ok {
		return "", fmt.Errorf("%s does not map to a string in %s map", id, DefaultVersions)
	}

	return version, nil
}

// Identity make Buildpack satisfy the Identifiable interface.
func (b Buildpack) Identity() (string, string) {
	return b.Info.Name, b.Info.Version
}

// IncludeFiles returns the include_files buildpack metadata.
func (b Buildpack) IncludeFiles() ([]string, error) {
	files, ok := b.Metadata["include_files"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var includes []string
	for _, candidate := range files {
		file, ok := candidate.(string)
		if !ok {
			return []string{}, fmt.Errorf("include_files is not an array of strings")
		}

		includes = append(includes, file)
	}

	return includes, nil
}

// PrePackage returns the pre_package buildpack metadata.
func (b Buildpack) PrePackage() (string, bool) {
	p, ok := b.Metadata["pre_package"].(string)
	return p, ok
}

// String makes Buildpack satisfy the Stringer interface.
func (b Buildpack) String() string {
	return fmt.Sprintf("Buildpack{ Buildpack: %s, CacheRoot: %s, logger: %s }",
		b.Buildpack, b.CacheRoot, b.logger)
}

// NewBuildpack creates a new instance of Buildpack from a specified buildpack.Buildpack.
func NewBuildpack(buildpack buildpack.Buildpack, logger logger.Logger) Buildpack {
	return Buildpack{buildpack, filepath.Join(buildpack.Root, cacheRoot), logger}
}
