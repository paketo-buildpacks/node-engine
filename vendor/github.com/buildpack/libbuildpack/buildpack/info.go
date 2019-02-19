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

// Info is information about the buildpack.
type Info struct {
	// ID is the globally unique identifier of the buildpack.
	ID string `toml:"id"`

	// Name is the human readable name of the buildpack.
	Name string `toml:"name"`

	// Version is the semver-compliant version of the buildpack.
	Version string `toml:"version"`
}

// String makes Info satisfy the Stringer interface.
func (i Info) String() string {
	return fmt.Sprintf("Info{ ID: %s, Name: %s, Version: %s }", i.ID, i.Name, i.Version)
}
