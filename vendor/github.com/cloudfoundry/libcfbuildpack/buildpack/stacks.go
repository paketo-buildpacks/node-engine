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

	"github.com/buildpack/libbuildpack/stack"
)

// Stacks is a collection of stack ids.
type Stacks []stack.Stack

// Validate ensures that there is at least one stack.
func (s Stacks) Validate() error {
	if len(s) == 0 {
		return fmt.Errorf("at least one stack is required")
	}

	return nil
}

func (s Stacks) contains(stack stack.Stack) bool {
	for _, v := range s {
		if v == stack {
			return true
		}
	}

	return false
}
