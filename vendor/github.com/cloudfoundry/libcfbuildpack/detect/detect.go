/*
 * Copyright 2019-2020 the original author or authors.
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
	"github.com/buildpack/libbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/cloudfoundry/libcfbuildpack/runner"
	"github.com/cloudfoundry/libcfbuildpack/services"
)

// Detect is an extension to libbuildpack.Detect that allows additional functionality to be added.
type Detect struct {
	detect.Detect

	// Buildpack represents the metadata associated with a buildpack.
	Buildpack buildpack.Buildpack

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	// Runner is used to run commands outside of the process.
	Runner runner.Runner

	// Services represents the services bound to the application.
	Services services.Services
}

// DefaultDetect creates a new instance of Detect using default values.  During initialization, all platform environment
// variables are set in the current process environment.
func DefaultDetect() (Detect, error) {
	d, err := detect.DefaultDetect()
	if err != nil {
		return Detect{}, err
	}

	logger := logger.Logger{Logger: d.Logger}
	buildpack := buildpack.NewBuildpack(d.Buildpack, logger)
	services := services.Services{Services: d.Services}

	return Detect{
		d,
		buildpack,
		logger,
		runner.CommandRunner{},
		services,
	}, nil
}
