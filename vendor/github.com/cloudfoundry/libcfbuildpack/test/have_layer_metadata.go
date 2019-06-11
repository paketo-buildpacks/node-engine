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
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/types"
)

// HaveLayerMetadata tests that a layer has a specific metadata configuration.
func HaveLayerMetadata(build bool, cache bool, launch bool) types.GomegaMatcher {
	return &haveLayerMetadataMatcher{
		build:  build,
		cache:  cache,
		launch: launch,
	}
}

type haveLayerMetadataMatcher struct {
	build  bool
	cache  bool
	launch bool
}

func (m *haveLayerMetadataMatcher) Match(actual interface{}) (bool, error) {
	metadata, err := m.getMetadata(actual)
	if err != nil {
		return false, err
	}

	if metadata["build"].(bool) != m.build {
		return false, nil
	}

	if metadata["cache"].(bool) != m.cache {
		return false, nil
	}

	if metadata["launch"].(bool) != m.launch {
		return false, nil
	}

	return true, nil
}

func (m *haveLayerMetadataMatcher) FailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf(
		"Expected\n\t%#v\nto have layer metadata\n\tbuild: %t, cache: %t, launch: %t\n"+
			"but found\n\tbuild: %t, cache: %t, launch: %t\n",
		actual, m.build, m.cache, m.launch,
		actualMetadata["build"], actualMetadata["cache"], actualMetadata["launch"])
}

func (m *haveLayerMetadataMatcher) NegatedFailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to have layer metadata\n\tbuild: %t, cache: %t, launch: %t\n"+
			"but found\n\tbuild: %t, cache: %t, launch: %t\n",
		actual, m.build, m.cache, m.launch,
		actualMetadata["build"], actualMetadata["cache"], actualMetadata["launch"])
}

func (m *haveLayerMetadataMatcher) path(actual interface{}) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Metadata")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HaveLayerMetadata matcher expects a layer")
	}

	return v.Interface().(string), nil
}

func (m *haveLayerMetadataMatcher) getMetadata(actual interface{}) (map[string]interface{}, error) {
	path, err := m.path(actual)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	if _, err := toml.DecodeFile(path, &metadata); err != nil {
		return nil, fmt.Errorf("failed to decode file: %s", err.Error())
	}

	return metadata, nil
}
