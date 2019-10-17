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

package test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/onsi/gomega/types"
)

// HaveAppendBuildEnvironment tests that a layer has an append build environment variable with the expected
// content.
func HaveAppendBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(fmt.Sprintf("%s.append", name), format, args...)
}

// HaveAppendLaunchEnvironment tests that a layer has an append launch environment variable with the expected
// content.
func HaveAppendLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(fmt.Sprintf("%s.append", name), format, args...)
}

// HaveAppendSharedEnvironment tests that a layer has an append shared environment variable with the expected
// content.
func HaveAppendSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(fmt.Sprintf("%s.append", name), format, args...)
}

// HaveAppendPathBuildEnvironment tests that a layer has an append path build environment variable with the expected
// content.
//
// Deprecated: Use HavePrependPathBuildEnvironment
func HaveAppendPathBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return HavePrependPathBuildEnvironment(name, format, args...)
}

// HaveAppendPathLaunchEnvironment tests that a layer has an append path launch environment variable with the expected
// content.
//
// Deprecated: Use HavePrependPathLaunchEnvironment
func HaveAppendPathLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return HavePrependPathLaunchEnvironment(name, format, args...)
}

// HaveAppendPathSharedEnvironment tests that a layer has an append path shared environment variable with the expected
// content.
//
// Deprecated: Use HavePrependPathSharedEnvironment
func HaveAppendPathSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return HavePrependPathSharedEnvironment(name, format, args...)
}

// HaveDefaultBuildEnvironment tests that a layer has a default build environment variable with the expected
// content.
func HaveDefaultBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(fmt.Sprintf("%s.default", name), format, args...)
}

// HaveDefaultLaunchEnvironment tests that a layer has a default launch environment variable with the expected
// content.
func HaveDefaultLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(fmt.Sprintf("%s.default", name), format, args...)
}

// HaveDefaultSharedEnvironment tests that a layer has a default shared environment variable with the expected
// content.
func HaveDefaultSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(fmt.Sprintf("%s.default", name), format, args...)
}

// HaveDelimiterBuildEnvironment tests that a layer has a delimiter for a build environment variable with the expected
// content.
func HaveDelimiterBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(fmt.Sprintf("%s.delim", name), format, args...)
}

// HaveDelimiterLaunchEnvironment tests that a layer has a delimiter for a launch environment variable with the expected
// content.
func HaveDelimiterLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(fmt.Sprintf("%s.delim", name), format, args...)
}

// HaveDelimiterSharedEnvironment tests that a layer has a delimiter for a shared environment variable with the expected
// content.
func HaveDelimiterSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(fmt.Sprintf("%s.delim", name), format, args...)
}

// HaveOverrideBuildEnvironment tests that a layer has an override build environment variable with the expected
// content.
func HaveOverrideBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(fmt.Sprintf("%s.override", name), format, args...)
}

// HaveOverrideLaunchEnvironment tests that a layer has an override launch environment variable with the expected
// content.
func HaveOverrideLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(fmt.Sprintf("%s.override", name), format, args...)
}

// HaveOverrideSharedEnvironment tests that a layer has an override shared environment variable with the expected
// content.
func HaveOverrideSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(fmt.Sprintf("%s.override", name), format, args...)
}

// HavePrependBuildEnvironment tests that a layer has a prepend build environment variable with the expected
// content.
func HavePrependBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(fmt.Sprintf("%s.prepend", name), format, args...)
}

// HavePrependLaunchEnvironment tests that a layer has a prepend launch environment variable with the expected
// content.
func HavePrependLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(fmt.Sprintf("%s.prepend", name), format, args...)
}

// HavePrependSharedEnvironment tests that a layer has a prepend shared environment variable with the expected
// content.
func HavePrependSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(fmt.Sprintf("%s.prepend", name), format, args...)
}

// HavePrependPathBuildEnvironment tests that a layer has a prepend path build environment variable with the expected
// content.
func HavePrependPathBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveBuildEnvironment(name, format, args...)
}

// HavePrependPathLaunchEnvironment tests that a layer has a prepend path launch environment variable with the expected
// content.
func HavePrependPathLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveLaunchEnvironment(name, format, args...)
}

// HavePrependPathSharedEnvironment tests that a layer has a prepend path shared environment variable with the expected
// content.
func HavePrependPathSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return haveSharedEnvironment(name, format, args...)
}

func haveBuildEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return &haveEnvironmentMatcher{
		filepath.Join("env.build", name),
		fmt.Sprintf(format, args...),
	}
}

func haveLaunchEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return &haveEnvironmentMatcher{
		filepath.Join("env.launch", name),
		fmt.Sprintf(format, args...),
	}
}

func haveSharedEnvironment(name string, format string, args ...interface{}) types.GomegaMatcher {
	return &haveEnvironmentMatcher{
		filepath.Join("env", name),
		fmt.Sprintf(format, args...),
	}
}

type haveEnvironmentMatcher struct {
	relativePath string
	expected     string
}

func (m *haveEnvironmentMatcher) Match(actual interface{}) (bool, error) {
	path, err := m.path(actual, m.relativePath)
	if err != nil {
		return false, err
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %s", err.Error())
	}

	return string(b) == m.expected, nil
}

func (m *haveEnvironmentMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto have layer environment %#v\n\t%#v", actual, m.relativePath, m.expected)
}

func (m *haveEnvironmentMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to have layer environment %#v\n\t%#v", actual, m.relativePath, m.expected)
}

func (m *haveEnvironmentMatcher) path(actual interface{}, relativePath string) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Root")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HaveEnvironment matcher expects a layer")
	}

	return filepath.Join(v.Interface().(string), relativePath), nil
}
