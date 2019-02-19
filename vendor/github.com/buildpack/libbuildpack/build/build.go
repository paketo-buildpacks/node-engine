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

package build

import (
	"fmt"

	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildpack"
	"github.com/buildpack/libbuildpack/buildplan"
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

	// BuildPlan represents dependencies contributed by previous builds.
	BuildPlan buildplan.BuildPlan

	// BuildPlanWriter is the writer used to write the BuildPlan in Pass().
	BuildPlanWriter buildplan.Writer

	// Layers represents the launch layers contributed by a buildpack.
	Layers layers.Layers

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	// Platform represents components contributed by the platform to the buildpack.
	Platform platform.Platform

	// Services represents the services bound to the application.
	Services services.Services

	// Stack is the stack currently available to the application.
	Stack stack.Stack
}

// Failure signals an unsuccessful build by exiting with a specified positive status code.
func (b Build) Failure(code int) int {
	b.Logger.Debug("Build failed. Exiting with %d.", code)
	return code
}

// String makes Build satisfy the Stringer interface.
func (b Build) String() string {
	return fmt.Sprintf("Build{ Application: %s, Buildpack: %s, BuildPlan: %s, BuildPlanWriter: %v, Layers: %s, Logger: %s, Platform: %s, Services: %s, Stack: %s }",
		b.Application, b.Buildpack, b.BuildPlan, b.BuildPlanWriter, b.Layers, b.Logger, b.Platform, b.Services, b.Stack)
}

// Success signals a successful build by exiting with a zero status code.
func (b Build) Success(buildPlan buildplan.BuildPlan) (int, error) {
	b.Logger.Debug("Build success. Exiting with %d.", SuccessStatusCode)

	if err := buildPlan.Write(b.BuildPlanWriter); err != nil {
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

	buildPlan := buildplan.BuildPlan{}
	if err := buildPlan.Init(); err != nil {
		return Build{}, err
	}

	buildPlanWriter := buildplan.DefaultWriter(3)

	layersRoot, err := internal.Argument(1)
	if err != nil {
		return Build{}, err
	}
	layers := layers.NewLayers(layersRoot, logger)

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

	return Build{
		application,
		buildpack,
		buildPlan,
		buildPlanWriter,
		layers,
		logger,
		platform,
		services,
		stack,
	}, nil
}
