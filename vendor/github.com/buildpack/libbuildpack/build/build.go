/*
 * Copyright 2018-2020 the original author or authors.
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
	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildpack"
	"github.com/buildpack/libbuildpack/buildpackplan"
	"github.com/buildpack/libbuildpack/internal"
	"github.com/buildpack/libbuildpack/layers"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpack/libbuildpack/platform"
	"github.com/buildpack/libbuildpack/services"
	"github.com/buildpack/libbuildpack/stack"
)

// SuccessStatusCode is the status code returned for success.
const SuccessStatusCode = 0

// Build represents all of the components available to a buildpack at build time.
type Build struct {
	// Application is the application being processed by the buildpack.
	Application application.Application

	// Buildpack represents the metadata associated with a buildpack.
	Buildpack buildpack.Buildpack

	// Layers represents the launch layers contributed by a buildpack.
	Layers layers.Layers

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	// Plans represents required contributions.
	Plans buildpackplan.Plans

	// Platform represents components contributed by the platform to the buildpack.
	Platform platform.Platform

	// Services represents the services bound to the application.
	Services services.Services

	// Stack is the stack currently available to the application.
	Stack stack.Stack

	// Writer is the writer used to write the build plan in Success().
	Writer buildpackplan.Writer
}

// Failure signals an unsuccessful build by exiting with a specified positive status code.
func (b Build) Failure(code int) int {
	b.Logger.Debug("Build failed. Exiting with %d.", code)
	return code
}

// Success signals a successful build by exiting with a zero status code.
func (b Build) Success(plans ...buildpackplan.Plan) (int, error) {
	b.Logger.Debug("Build success. Exiting with %d.", SuccessStatusCode)

	if err := b.Writer(buildpackplan.Plans{Entries: plans}); err != nil {
		return -1, err
	}

	return SuccessStatusCode, nil
}

// DefaultBuild creates a new instance of Build using default values.
func DefaultBuild() (Build, error) {
	platformRoot, err := internal.Argument(2)
	if err != nil {
		return Build{}, err
	}

	logger, err := logger.DefaultLogger(platformRoot)
	if err != nil {
		return Build{}, nil
	}

	application, err := application.DefaultApplication(logger)
	if err != nil {
		return Build{}, err
	}

	buildpack, err := buildpack.DefaultBuildpack(logger)
	if err != nil {
		return Build{}, err
	}

	plan, err := internal.Argument(3)
	if err != nil {
		return Build{}, err
	}

	layersRoot, err := internal.Argument(1)
	if err != nil {
		return Build{}, err
	}
	layers := layers.NewLayers(layersRoot, logger)

	plans, err := buildpackplan.DefaultPlans(plan, logger)
	if err != nil {
		return Build{}, err
	}

	platform, err := platform.DefaultPlatform(platformRoot, logger)
	if err != nil {
		return Build{}, err
	}

	services, err := services.DefaultServices(platform, logger)
	if err != nil {
		return Build{}, err
	}

	stack, err := stack.DefaultStack(logger)
	if err != nil {
		return Build{}, err
	}

	writer := buildpackplan.DefaultWriter(3)

	return Build{
		application,
		buildpack,
		layers,
		logger,
		plans,
		platform,
		services,
		stack,
		writer,
	}, nil
}
