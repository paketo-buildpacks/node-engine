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

import (
	"fmt"
)

// License represents a license that a Dependency is distributed under.  At least one of Name or URI MUST be specified.
type License struct {
	// Type is the type of the license.  This is typically the SPDX short identifier.
	Type string `mapstruct:"type" toml:"type"`

	// URI is the location where the license can be found.
	URI string `mapstruct:"uri" toml:"uri"`
}

// Validate ensures that license has at least one of type or uri
func (l License) Validate() error {
	if "" == l.Type && "" == l.URI {
		return fmt.Errorf("license must have at least one of type or uri")
	}

	return nil
}
