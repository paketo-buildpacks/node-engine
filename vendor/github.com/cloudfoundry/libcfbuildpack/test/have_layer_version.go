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
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/types"
)

// HaveLayerVersion tests that a layer has a specific version.
func HaveLayerVersion(version string) types.GomegaMatcher {
	return &haveLayerVersion{
		version: version,
	}
}

type haveLayerVersion struct {
	version string
}

func (m *haveLayerVersion) Match(actual interface{}) (bool, error) {
	path, err := m.path(actual)
	if err != nil {
		return false, err
	}

	layer := struct {
		Metadata struct {
			Version string `toml:"version"`
		} `toml:"metadata"`
	}{}

	if _, err := toml.DecodeFile(path, &layer); err != nil {
		return false, fmt.Errorf("failed to decode file: %s", err.Error())
	}

	if layer.Metadata.Version != m.version {
		return false, nil
	}

	return true, nil
}

func (m *haveLayerVersion) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto have layer metadata\n\tversion: %s",
		actual, m.version)
}

func (m *haveLayerVersion) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to have layer metadata\n\tversion: %s",
		actual, m.version)
}

func (m *haveLayerVersion) path(actual interface{}) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Metadata")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HaveLayerVersion matcher expects a layer")
	}

	return v.Interface().(string), nil
}
