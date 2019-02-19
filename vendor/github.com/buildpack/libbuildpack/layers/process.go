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

package layers

import (
	"fmt"
)

// Processes is a collection of Process instances.
type Processes []Process

// Process represents metadata about a type of command that can be run.
type Process struct {
	// Type is the type of the process.
	Type string `toml:"type"`

	// Command is the command of the process.
	Command string `toml:"command"`
}

// String makes Process satisfy the Stringer interface.
func (p Process) String() string {
	return fmt.Sprintf("Process{ Type: %s, Command: %s }", p.Type, p.Command)
}
