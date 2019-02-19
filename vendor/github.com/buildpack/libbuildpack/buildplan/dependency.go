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

package buildplan

import (
	"fmt"
)

// Dependency represents a dependency in a build.
type Dependency struct {
	// Version is the optional dependency version.
	Version string `toml:"version"`

	// Metadata is additional metadata attached to the dependency.
	Metadata Metadata `toml:"metadata"`
}

// String makes Dependency satisfy the Stringer interface.
func (d Dependency) String() string {
	return fmt.Sprintf("Dependency{ Version: %s, Metadata: %s }", d.Version, d.Metadata)
}

// Metadata is additional metadata attached to a dependency.
type Metadata = map[string]interface{}
