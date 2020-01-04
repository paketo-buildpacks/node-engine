/*
 * Copyright 2018-2020 the original author or authors.
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

package buildpackplan

import (
	"strings"
)

// ShallowMerge merges two Plans together.  Declared versions are combined with a comma delimiter and metadata is
// combined with the values for b taking priority over the values of a when keys are duplicated.
func ShallowMerge(a, b Plan) (Plan, error) {
	m := a

	m.Version = mergeVersion(a.Version, b.Version)
	m.Metadata = mergeMetadata(a.Metadata, b.Metadata)

	return m, nil
}

func mergeMetadata(a, b Metadata) Metadata {
	if a == nil && b == nil {
		return nil
	}

	m := make(Metadata)

	for k, v := range a {
		m[k] = v
	}

	for k, v := range b {
		m[k] = v
	}

	return m
}

func mergeVersion(a, b string) string {
	var v []string

	if a != "" {
		v = append(v, a)
	}

	if b != "" {
		v = append(v, b)
	}

	return strings.Join(v, ",")
}
