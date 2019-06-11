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
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/onsi/gomega/types"
)

// HaveApplicationMetadata tests that an application metadata has expected content.
func HaveApplicationMetadata(expected layers.Metadata) types.GomegaMatcher {
	return &haveApplicationMetadataMatcher{
		expected: expected,
	}
}

type haveApplicationMetadataMatcher struct {
	expected layers.Metadata
}

func (m *haveApplicationMetadataMatcher) Match(actual interface{}) (bool, error) {
	metadata, err := m.getMetadata(actual)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(metadata, m.expected), nil
}

func (m *haveApplicationMetadataMatcher) FailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Expected\n\t%#v\nto have application metadata\n\t%#v\n"+
		"but found\n\t%#v\n", actual, m.expected, actualMetadata)
}

func (m *haveApplicationMetadataMatcher) NegatedFailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("Expected\n\t%#v\nnot to have application metadata\n\t%#v\n"+
		"but found\n\t%#v\n", actual, m.expected, actualMetadata)
}

func (m *haveApplicationMetadataMatcher) path(actual interface{}) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Root")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HaveApplicationMetadata matcher expects a layers")
	}

	return filepath.Join(v.Interface().(string), "app.toml"), nil
}

func (m *haveApplicationMetadataMatcher) getMetadata(actual interface{}) (layers.Metadata, error) {
	path, err := m.path(actual)
	if err != nil {
		return layers.Metadata{}, err
	}

	var metadata layers.Metadata
	if _, err := toml.DecodeFile(path, &metadata); err != nil {
		return layers.Metadata{}, fmt.Errorf("failed to decode file: %s", err.Error())
	}

	return metadata, nil
}
