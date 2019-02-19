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

// Licenses is a collection of licenses
type Licenses []License

// Validate ensures that there is at least one license and all licenses are valid
func (l Licenses) Validate() error {
	if len(l) == 0 {
		return fmt.Errorf("at least one license is required")
	}

	for _, license := range l {
		if err := license.Validate(); err != nil {
			return err
		}
	}

	return nil
}
