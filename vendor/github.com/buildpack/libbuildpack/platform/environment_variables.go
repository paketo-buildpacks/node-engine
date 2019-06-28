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

package platform

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/logger"
)

// EnvironmentVariables is a collection of environment variables provided by the Platform.
type EnvironmentVariables map[string]string

// SetAll sets all of the environment variable content in the current process environment.
func (e EnvironmentVariables) SetAll() error {
	for key, value := range e {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}

func environmentVariables(root string, logger logger.Logger) (EnvironmentVariables, error) {
	files, err := filepath.Glob(filepath.Join(root, "env", "*"))
	if err != nil {
		return nil, err
	}

	e := make(EnvironmentVariables)

	for _, file := range files {
		value, err := value(file)
		if err != nil {
			return nil, err
		}

		e[filepath.Base(file)] = value
	}

	logger.Debug("Platform environment variables: %s", e)
	return e, nil
}

func value(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
