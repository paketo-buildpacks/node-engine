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

package buildpackplan

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/logger"
)

// Plan represents a contractual buildpack plan.
type Plan struct {
	// Name represents the name of the plan.
	Name string `toml:"name"`

	// Version represents the version of the plan.  Optional.
	Version string `toml:"version,omitempty"`

	// Metadata represents the metadata of the plan.  Optional.
	Metadata Metadata `toml:"metadata,omitempty"`
}

// Metadata is the metadata of the plan.
type Metadata map[string]interface{}

// Plans is a collection of Plan's in marshalable form.
type Plans struct {
	// Entries represents all of the buildpack plans.
	Entries []Plan `toml:"entries,omitempty"`
}

// DefaultPlans creates a new instance of Plans, unmarshalling it from a TOML file.
func DefaultPlans(path string, logger logger.Logger) (Plans, error) {
	in, err := os.Open(path)
	if err != nil {
		return Plans{}, err
	}
	defer in.Close()

	b := Plans{}

	if _, err := toml.DecodeReader(in, &b); err != nil {
		return Plans{}, err
	}

	logger.Debug("Plans: %#v", b)
	return b, nil
}
