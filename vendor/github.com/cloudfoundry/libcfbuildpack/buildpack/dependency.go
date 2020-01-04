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

package buildpack

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// Dependency represents a buildpack dependency.
type Dependency struct {
	// ID is the dependency ID.
	ID string `mapstruct:"id" toml:"id"`

	// Name is the dependency ID.
	Name string `mapstruct:"name" toml:"name"`

	// Version is the dependency version.
	Version Version `mapstruct:"version" toml:"version"`

	// URI is the dependency URI.
	URI string `mapstruct:"uri" toml:"uri"`

	// SHA256 is the hash of the dependency.
	SHA256 string `mapstruct:"sha256" toml:"sha256"`

	// Stacks are the stacks the dependency is compatible with.
	Stacks Stacks `mapstruct:"stacks" toml:"stacks"`

	// Licenses are the stacks the dependency is distributed under.
	Licenses Licenses `mapstruct:"licenses" toml:"licenses"`
}

// NewDependency makes a Dependency from a generic map describing a Dependency
func NewDependency(dep map[string]interface{}) (Dependency, error) {
	var d Dependency

	config := mapstructure.DecoderConfig{
		DecodeHook: unmarshalText,
		Result:     &d,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return Dependency{}, err
	}

	if err := decoder.Decode(dep); err != nil {
		return Dependency{}, err
	}

	return d, nil
}

// Identity make Buildpack satisfy the Identifiable interface.
func (d Dependency) Identity() (string, string) {
	if d.Version.Version != nil {
		return d.Name, d.Version.Version.Original()
	}

	return d.Name, ""
}

// Validate ensures that the dependency is valid.
func (d Dependency) Validate() error {
	if "" == d.ID {
		return fmt.Errorf("id is required")
	}

	if "" == d.Name {
		return fmt.Errorf("name is required")
	}

	if (Version{} == d.Version) {
		return fmt.Errorf("version is required")
	}

	if "" == d.URI {
		return fmt.Errorf("uri is required")
	}

	if "" == d.SHA256 {
		return fmt.Errorf("sha256 is required")
	}

	if err := d.Stacks.Validate(); err != nil {
		return err
	}

	if err := d.Licenses.Validate(); err != nil {
		return err
	}

	return nil
}
