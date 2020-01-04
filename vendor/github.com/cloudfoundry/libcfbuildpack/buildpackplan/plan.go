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

package buildpackplan

import (
	"reflect"

	"github.com/buildpack/libbuildpack/buildpackplan"
)

// Plan represents a contractual buildpack plan.
type Plans struct {
	buildpackplan.Plans
}

type MergeFunc func(a, b Plan) (Plan, error)

// Get returns a collection of Plan's that have a given name.
func (p Plans) Get(name string) []Plan {
	var ps []Plan

	for _, p := range p.Entries {
		if p.Name == name {
			ps = append(ps, p)
		}
	}

	return ps
}

// GetMerged returns a single Plan that is a merged version of all of the Plan's that have a given name. A merge
// function is used to describe how two entries are merged together.  Returns true if any matching Plan's were found,
// false otherwise.
func (p Plans) GetMerged(name string, merge MergeFunc) (Plan, bool, error) {
	m := Plan{}
	var err error

	for _, p := range p.Entries {
		if p.Name == name {
			if reflect.DeepEqual(m, Plan{}) {
				m = Plan(p)
			} else {
				m, err = merge(m, p)
				if err != nil {
					return Plan{}, false, err
				}
			}
		}
	}

	return m, !reflect.DeepEqual(m, Plan{}), nil
}

// GetShallowMerged returns a single Plan that is a merged version of all of the Plan's that have a given name.  Merging
// is accomplished with the ShallowMerge function.  Returns true if any matching Plan's were found, false otherwise.
func (p Plans) GetShallowMerged(name string) (Plan, bool, error) {
	return p.GetMerged(name, ShallowMerge)
}

// GetPriorityMerged returns a single Plan that is a merged version of all of the Plan's that have a given name.  Merging
// is accomplished with the PriorityMerge function.  Returns true if any matching Plan's were found, false otherwise.
func (p Plans) GetPriorityMerged(name string, priorities map[interface{}]int) (Plan, bool, error) {
	return p.GetMerged(name, PriorityMerge(priorities))
}

// Has returns whether there is a Plan with a given name.
func (p Plans) Has(name string) bool {
	for _, p := range p.Entries {
		if p.Name == name {
			return true
		}
	}

	return false
}
