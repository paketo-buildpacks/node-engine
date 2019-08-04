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

package buildplan

// Plan represents a contractual build plan.
type Plan struct {
	// Provided represents the dependencies provided by a buildpack. Optional.
	Provides []Provided `toml:"provides,omitempty"`

	// Required represents the dependencies required by a buildpack.  Optional.
	Requires []Required `toml:"requires,omitempty"`
}

// Provided represents a dependency provided by a buildpack.
type Provided struct {
	// Name represents the name of a dependency.
	Name string `toml:"name"`
}

// Required represents a dependency required by a buildpack.
type Required struct {
	// Name represents the name of a dependency.
	Name string `toml:"name"`

	// Version represents the version of a dependency.  Optional.
	Version string `toml:"version,omitempty"`

	// Metadata represents the metadata of a dependency.  Optional.
	Metadata Metadata `toml:"metadata,omitempty"`
}

// Metadata is the metadata of a dependency.
type Metadata map[string]interface{}

// Plans is a collection of Plan's in marshalable form.
type Plans struct {
	// Plan is the primary Plan.
	Plan

	// Or are additional Plans. Optional.
	Or []Plan `toml:"or,omitempty"`
}
