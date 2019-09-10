/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package build

import (
	"github.com/buildpack/libbuildpack/build"
	bp "github.com/buildpack/libbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/runner"
	"github.com/cloudfoundry/libcfbuildpack/services"
)

// Build is an extension to libbuildpack.Build that allows additional functionality to be added.
type Build struct {
	build.Build

	// Buildpack represents the metadata associated with a buildpack.
	Buildpack buildpack.Buildpack

	// Layers represents the launch layers contributed by a buildpack.
	Layers layers.Layers

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	// Plans represents required contributions.
	Plans buildpackplan.Plans

	// Runner is used to run commands outside of the process.
	Runner runner.Runner

	// Services represents the services bound to the application.
	Services services.Services
}

// Success signals a successful build by exiting with a zero status code.  Combines specied build plan with build
// plan entries for all contributed dependencies.
func (b Build) Success(plans ...buildpackplan.Plan) (int, error) {

	code, err := b.Build.Success(append(b.Layers.Plans.Entries, plans...)...)
	if err != nil {
		return code, err
	}

	if err := b.Layers.TouchedLayers.Cleanup(); err != nil {
		return -1, err
	}

	return code, nil
}

// DefaultBuild creates a new instance of Build using default values.  During initialization, all platform environment
// // variables are set in the current process environment.
func DefaultBuild() (Build, error) {
	b, err := build.DefaultBuild()
	if err != nil {
		return Build{}, err
	}

	logger := logger.Logger{Logger: b.Logger}
	buildpack := buildpack.NewBuildpack(b.Buildpack, logger)
	layers := layers.NewLayers(b.Layers, bp.NewLayers(buildpack.CacheRoot, b.Logger), buildpack, logger)
	plans := buildpackplan.Plans{Plans: b.Plans}
	services := services.Services{Services: b.Services}

	return Build{
		b,
		buildpack,
		layers,
		logger,
		plans,
		runner.CommandRunner{},
		services,
	}, nil
}
