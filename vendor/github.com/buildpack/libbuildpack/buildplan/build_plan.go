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

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/internal"
)

// BuildPlan represents the dependencies contributed by a build.  Note that you may need to call Init() to load contents
// from os.Stdin.
type BuildPlan map[string]Dependency

// Init initializes the BuildPlan by reading os.Stdin.  Will block until os.Stdin is closed.
func (b BuildPlan) Init() error {
	if _, err := toml.DecodeReader(os.Stdin, &b); err != nil {
		return err
	}

	return nil
}

// Merge performs a shallow merge of the entries in passed BuildPlans into this.
func (b BuildPlan) Merge(buildPlans ...BuildPlan) {
	for _, bp := range buildPlans {
		for k, v := range bp {
			b[k] = v
		}
	}
}

// Write writes the build plan.
func (b BuildPlan) Write(writer Writer) error {
	return writer(b)
}

// Writer is a function write writes the contents of a BuildPlan
type Writer func(buildPlan BuildPlan) error

// DefaultWriter writes the build plan to a collection of files rooted at os.Args[<INDEX>].
func DefaultWriter(index int) Writer {
	return func(buildPlan BuildPlan) error {
		path, err := internal.Argument(index)
		if err != nil {
			return err
		}

		return internal.WriteTomlFile(path, 0644, buildPlan)
	}
}
