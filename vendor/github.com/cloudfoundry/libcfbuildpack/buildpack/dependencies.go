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
	"sort"

	"github.com/Masterminds/semver"
	"github.com/buildpack/libbuildpack/stack"
)

// Dependencies is a collection of Dependency instances.
type Dependencies []Dependency

// Best returns the best (latest version) dependency within a collection of Dependencies.  The candidate set is first
// filtered by id, version, and stack, then the remaining candidates are sorted for the best result.  If the
// versionConstraint is not specified (""), then the latest wildcard ("*") is used.
func (d Dependencies) Best(id string, versionConstraint string, stack stack.Stack) (Dependency, error) {
	var candidates Dependencies

	vc := versionConstraint
	if vc == "" {
		vc = "*"
	}

	constraint, err := semver.NewConstraint(vc)
	if err != nil {
		return Dependency{}, err
	}

	for _, c := range d {
		if c.ID == id && constraint.Check(c.Version.Version) && c.Stacks.contains(stack) {
			candidates = append(candidates, c)
		}
	}

	if len(candidates) == 0 {
		return Dependency{}, fmt.Errorf("no valid dependencies for %s, %s, and %s in %s", id, vc, stack, d)
	}

	sort.Slice(candidates, func(i int, j int) bool {
		return d[i].Version.LessThan(d[j].Version.Version)
	})

	return candidates[len(candidates)-1], nil
}

// Has indicates whether the collection of dependencies has any dependency of a specific id.  This is used primarily to
// determine whether an optional dependency exists, before calling Best() which would throw an error if one did not.
func (d Dependencies) Has(id string) bool {
	for _, c := range d {
		if c.ID == id {
			return true
		}
	}

	return false
}
