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

package test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	bpLayers "github.com/buildpack/libbuildpack/layers"
	bpServices "github.com/buildpack/libbuildpack/services"
	"github.com/buildpack/libbuildpack/stack"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/internal"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/services"
)

// BuildFactory is a factory for creating a test Build.
type BuildFactory struct {
	// Build is the configured build to use.
	Build build.Build

	// Home is the home directory to use.
	Home string

	// Output is the BuildPlan output at termination.
	Output buildplan.BuildPlan

	// Runner is the used to capture commands executed outside the process.
	Runner *Runner

	t *testing.T
}

// AddBuildPlan adds an entry to a build plan.
func (f *BuildFactory) AddBuildPlan(name string, dependency buildplan.Dependency) {
	f.t.Helper()

	if f.Build.BuildPlan == nil {
		f.Build.BuildPlan = make(buildplan.BuildPlan)
	}

	f.Build.BuildPlan[name] = dependency
}

// AddDependency adds a dependency with version 1.0 to the buildpack metadata and copies a fixture into a cached
// dependency layer.
func (f *BuildFactory) AddDependency(id string, fixturePath string) {
	f.t.Helper()
	f.AddDependencyWithVersion(id, "1.0", fixturePath)
}

// AddDependencyWithVersion adds a dependency to the buildpack metadata and copies a fixture into a cached dependency
// layer
func (f *BuildFactory) AddDependencyWithVersion(id string, version string, fixturePath string) {
	f.t.Helper()

	d := f.newDependency(id, version, filepath.Base(fixturePath))
	f.cacheFixture(d, fixturePath)
	f.addDependency(d)
}

// SetDefaultVersion sets a default dependency version in the buildpack metadata
func (f *BuildFactory) SetDefaultVersion(id, version string) {
	f.t.Helper()

	if f.Build.Buildpack.Metadata == nil {
		f.Build.Buildpack.Metadata = make(buildpack.Metadata)
	}

	if _, ok := f.Build.Buildpack.Metadata[buildpack.DefaultVersions]; !ok {
		f.Build.Buildpack.Metadata[buildpack.DefaultVersions] = map[string]interface{}{}
	}

	metadata := f.Build.Buildpack.Metadata
	metadata[buildpack.DefaultVersions].(map[string]interface{})[id] = version
}

// AddService adds an entry to the collection of services.
func (f *BuildFactory) AddService(name string, credentials services.Credentials, tags ...string) {
	f.t.Helper()

	f.Build.Services.Services = append(f.Build.Services.Services, bpServices.Service{
		BindingName: name,
		Credentials: credentials,
		Tags:        tags,
	})
}

func (f *BuildFactory) addDependency(dependency buildpack.Dependency) {
	f.t.Helper()

	if f.Build.Buildpack.Metadata == nil {
		f.Build.Buildpack.Metadata = make(buildpack.Metadata)
	}

	if _, ok := f.Build.Buildpack.Metadata[buildpack.DependenciesMetadata]; !ok {
		f.Build.Buildpack.Metadata[buildpack.DependenciesMetadata] = make([]map[string]interface{}, 0)
	}

	metadata := f.Build.Buildpack.Metadata
	dependencies := metadata[buildpack.DependenciesMetadata].([]map[string]interface{})

	var stacks []interface{}
	for _, stack := range dependency.Stacks {
		stacks = append(stacks, stack)
	}

	var licenses []map[string]interface{}
	for _, license := range dependency.Licenses {
		licenses = append(licenses, map[string]interface{}{
			"type": license.Type,
			"uri":  license.URI,
		})
	}

	metadata[buildpack.DependenciesMetadata] = append(dependencies, map[string]interface{}{
		"id":       dependency.ID,
		"name":     dependency.Name,
		"version":  dependency.Version.Version.Original(),
		"uri":      dependency.URI,
		"sha256":   dependency.SHA256,
		"stacks":   stacks,
		"licenses": licenses,
	})
}

func (f *BuildFactory) cacheFixture(dependency buildpack.Dependency, fixturePath string) {
	f.t.Helper()

	l := f.Build.Layers.Layer(dependency.SHA256)
	if err := helper.CopyFile(fixturePath, filepath.Join(l.Root, dependency.Name)); err != nil {
		f.t.Fatal(err)
	}

	if err := internal.WriteTomlFile(l.Metadata, 0644, map[string]interface{}{"metadata": dependency}); err != nil {
		f.t.Fatal(err)
	}
}

func (f *BuildFactory) newDependency(id string, version string, name string) buildpack.Dependency {
	f.t.Helper()

	return buildpack.Dependency{
		ID:      id,
		Name:    name,
		Version: internal.NewTestVersion(f.t, version),
		SHA256:  hex.EncodeToString(sha256.New().Sum([]byte(id))),
		URI:     fmt.Sprintf("http://localhost/%s", name),
		Stacks:  buildpack.Stacks{f.Build.Stack},
	}
}

// NewBuildFactory creates a new instance of BuildFactory.
func NewBuildFactory(t *testing.T) *BuildFactory {
	t.Helper()

	root := ScratchDir(t, "build")
	runner := &Runner{}

	f := BuildFactory{Home: filepath.Join(root, "home"), Runner: runner, t: t}

	f.Build.Application.Root = filepath.Join(root, "application")
	if err := os.MkdirAll(f.Build.Application.Root, 0755); err != nil {
		t.Fatal(err)
	}
	f.Build.Buildpack.Info.Version = "1.0"
	f.Build.Buildpack.Root = filepath.Join(root, "buildpack")
	f.Build.BuildPlanWriter = func(buildPlan buildplan.BuildPlan) error {
		f.Output = buildPlan
		return nil
	}
	f.Build.Layers = layers.NewLayers(
		bpLayers.Layers{Root: filepath.Join(root, "layers")},
		bpLayers.Layers{Root: filepath.Join(root, "buildpack-cache")}, f.Build.Buildpack, logger.Logger{})
	f.Build.Platform.Root = filepath.Join(root, "platform")
	f.Build.Runner = runner
	f.Build.Services = services.Services{Services: bpServices.Services{}}
	f.Build.Stack = stack.Stack("test-stack")

	return &f
}
