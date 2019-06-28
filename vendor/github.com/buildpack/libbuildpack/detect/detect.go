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

package detect

import (
	"github.com/buildpack/libbuildpack/application"
	"github.com/buildpack/libbuildpack/buildpack"
	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/buildpack/libbuildpack/internal"
	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpack/libbuildpack/platform"
	"github.com/buildpack/libbuildpack/services"
	"github.com/buildpack/libbuildpack/stack"
)

const (
	// FailStatusCode is the status code returned for fail.
	FailStatusCode = 100

	// PassStatusCode is the status code returned for pass.
	PassStatusCode = 0
)

// Detect represents all of the components available to a buildpack at detect time.
type Detect struct {
	// Application is the application being processed by the buildpack.
	Application application.Application

	// Buildpack represents the metadata associated with a buildpack.
	Buildpack buildpack.Buildpack

	// BuildPlan represents dependencies contributed by previous builds.
	BuildPlan buildplan.BuildPlan

	// BuildPlanWriter is the writer used to write the BuildPlan in Pass().
	BuildPlanWriter buildplan.Writer

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	// Platform represents components contributed by the platform to the buildpack.
	Platform platform.Platform

	// Services represents the services bound to the application.
	Services services.Services

	// Stack is the stack currently available to the application.
	Stack stack.Stack
}

// Error signals an error during detection by exiting with a specified non-zero, non-100 status code.
func (d Detect) Error(code int) int {
	d.Logger.Debug("Detection produced an error. Exiting with %d.", code)
	return code
}

// Fail signals an unsuccessful detection by exiting with a 100 status code.
func (d Detect) Fail() int {
	d.Logger.Debug("Detection failed. Exiting with %d.", FailStatusCode)
	return FailStatusCode
}

// Pass signals a successful detection by exiting with a 0 status code.
func (d Detect) Pass(buildPlan buildplan.BuildPlan) (int, error) {
	d.Logger.Debug("Detection passed. Exiting with %d.", PassStatusCode)

	if err := buildPlan.Write(d.BuildPlanWriter); err != nil {
		return -1, err
	}

	return PassStatusCode, nil
}

// DefaultDetect creates a new instance of Detect using default values.
func DefaultDetect() (Detect, error) {
	platformRoot, err := internal.Argument(1)
	if err != nil {
		return Detect{}, err
	}

	logger, err := logger.DefaultLogger(platformRoot)
	if err != nil {
		return Detect{}, err
	}

	application, err := application.DefaultApplication(logger)
	if err != nil {
		return Detect{}, err
	}

	buildpack, err := buildpack.DefaultBuildpack(logger)
	if err != nil {
		return Detect{}, err
	}

	buildPlan := buildplan.BuildPlan{}

	buildPlanWriter := buildplan.DefaultWriter(2)

	platform, err := platform.DefaultPlatform(platformRoot, logger)
	if err != nil {
		return Detect{}, err
	}

	services, err := services.DefaultServices(platform, logger)
	if err != nil {
		return Detect{}, err
	}

	stack, err := stack.DefaultStack(logger)
	if err != nil {
		return Detect{}, err
	}

	return Detect{
		application,
		buildpack,
		buildPlan,
		buildPlanWriter,
		logger,
		platform,
		services,
		stack,
	}, nil
}
