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

package buildpack

import (
	"fmt"
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

// Identity make Buildpack satisfy the Identifiable interface.
func (d Dependency) Identity() (string, string) {
	if d.Version.Version != nil {
		return d.Name, d.Version.Version.Original()
	}

	return d.Name, ""
}

// String makes Dependency satisfy the Stringer interface.
func (d Dependency) String() string {
	return fmt.Sprintf("Dependency{ ID: %s, Name: %s, Version: %s, URI: %s, SHA256: %s, Stacks: %s, Licenses: %s }",
		d.ID, d.Name, d.Version, d.URI, d.SHA256, d.Stacks, d.Licenses)
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
