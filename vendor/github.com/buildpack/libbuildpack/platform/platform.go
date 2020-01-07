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

package platform

import (
	"github.com/buildpack/libbuildpack/internal"
	"github.com/buildpack/libbuildpack/logger"
)

// Platform represents the platform contributions for an application.
type Platform struct {
	// Root is the path to the root directory for the platform contributions.
	Root string

	// EnvironmentVariables is the collection of environment variables contributed by the platform.
	EnvironmentVariables EnvironmentVariables

	logger logger.Logger
}

// DefaultPlatform creates a new instance of Platform.
func DefaultPlatform(root string, logger logger.Logger) (Platform, error) {
	if logger.IsDebugEnabled() {
		contents, err := internal.DirectoryContents(root)
		if err != nil {
			return Platform{}, err
		}
		logger.Debug("Platform contents: %s", contents)
	}

	environmentVariables, err := environmentVariables(root, logger)
	if err != nil {
		return Platform{}, err
	}

	return Platform{root, environmentVariables, logger}, err
}
