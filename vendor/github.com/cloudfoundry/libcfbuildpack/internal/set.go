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

package internal

// Set represents the mathematical type set.
type Set struct {
	contents map[interface{}]struct{}
}

// Add adds an element to the set.
func (s Set) Add(v interface{}) {
	s.contents[v] = struct{}{}
}

// Contains returns whether the set contains an item.
func (s Set) Contains(v interface{}) bool {
	_, ok := s.contents[v]
	return ok
}

// Difference provides the logical difference between this set and another.  Only elements that exist in this set and
// not the other will be in the resulting set.
func (s Set) Difference(other Set) Set {
	d := NewSet()

	for v := range s.Iterator() {
		if !other.Contains(v) {
			d.Add(v)
		}
	}

	return d
}

// Iterator is a type that range-able.
type Iterator <-chan interface{}

// Iterator returns the values to be ranged over.
func (s Set) Iterator() Iterator {
	ch := make(chan interface{})

	go func() {
		defer close(ch)

		for k, _ := range s.contents {
			ch <- k
		}
	}()

	return ch
}

// Size returns the number of elements in the set.
func (s Set) Size() int {
	return len(s.contents)
}

// NewSet creates an initialized and empty Set.
func NewSet() Set {
	return Set{make(map[interface{}]struct{})}
}
