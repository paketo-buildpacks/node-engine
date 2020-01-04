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

package test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/onsi/gomega/types"
)

// HaveProfile tests that a layer has a profile.d file with the expected content.
func HaveProfile(name string, format string, args ...interface{}) types.GomegaMatcher {
	return &haveProfileMatcher{
		name,
		fmt.Sprintf(format, args...),
	}
}

type haveProfileMatcher struct {
	name     string
	expected string
}

func (m *haveProfileMatcher) Match(actual interface{}) (bool, error) {
	path, err := m.path(actual, m.name)
	if err != nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %s", err.Error())
	}

	return string(b) == m.expected, nil
}

func (m *haveProfileMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto have layer profile %#v\n\t%#v", actual, m.name, m.expected)
}

func (m *haveProfileMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to have layer profile %#v\n\t%#v", actual, m.name, m.expected)
}

func (m *haveProfileMatcher) path(actual interface{}, name string) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Root")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HaveProfile matcher expects a layer")
	}

	return filepath.Join(v.Interface().(string), "profile.d", name), nil
}
