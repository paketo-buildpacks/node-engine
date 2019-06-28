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

package buildpack

// Stack represents metadata about the stacks associated with the buildpack.
type Stack struct {
	// ID is the globally unique identifier of the stack.
	ID string `toml:"id"`

	// BuildImages are the suggested sources for stacks if the platform is unaware of the stack id.
	BuildImages BuildImages `toml:"build-images"`

	// RunImages are the suggested sources for stacks if the platform is unaware of the stack id.
	RunImages RunImages `toml:"run-images"`
}

// BuildImages is a collection of BuildImages.
type BuildImages []BuildImage

// BuildImage is the build image source for a particular stack id.
type BuildImage string

// RunImages is a collection of RunImages.
type RunImages []RunImage

// RunImage is the run image source for a particular stack id.
type RunImage string
