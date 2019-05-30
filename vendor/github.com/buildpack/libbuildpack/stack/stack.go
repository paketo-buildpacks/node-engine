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

package stack

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/logger"
)

type Stack string

// DefaultStack creates a new instance of Stack, extracting the name from the CNB_STACK_ID environment variable.
func DefaultStack(logger logger.Logger) (Stack, error) {
	s, ok := os.LookupEnv("CNB_STACK_ID")
	if !ok {
		return "", fmt.Errorf("CNB_STACK_ID not set")
	}

	stack := Stack(s)
	logger.Debug("Stack: %s", stack)

	return stack, nil
}
