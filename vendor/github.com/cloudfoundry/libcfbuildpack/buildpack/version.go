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
	"reflect"

	"github.com/Masterminds/semver"
)

// Version is an extension to semver.Version to make it marshalable.
type Version struct {
	*semver.Version
}

// MarshalText makes Version satisfy the encoding.TextMarshaler interface.
func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.Version.Original()), nil
}

// UnmarshalText makes Version satisfy the encoding.TextUnmarshaler interface.
func (v *Version) UnmarshalText(text []byte) error {
	s := string(text)

	w, err := semver.NewVersion(s)
	if err != nil {
		return fmt.Errorf("invalid semantic version %s", s)
	}

	v.Version = w
	return nil
}

func unmarshalText(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() != reflect.String {
		return data, nil
	}

	if to != reflect.TypeOf(Version{}) {
		return data, nil
	}

	w, err := semver.NewVersion(data.(string))
	if err != nil {
		return nil, fmt.Errorf("invalid semantic version %s", data)
	}

	return Version{w}, nil
}
