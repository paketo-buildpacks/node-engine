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
	"github.com/onsi/gomega/types"
)

// HavePersistentMetadata tests that an persistent metadata has expected content.
func HavePersistentMetadata(expected interface{}) types.GomegaMatcher {
	return &havePersistentMetadataMatcher{
		expected: expected,
	}
}

type havePersistentMetadataMatcher struct {
	expected interface{}
}

func (m *havePersistentMetadataMatcher) Match(actual interface{}) (bool, error) {
	metadata, err := m.getMetadata(actual)
	if err != nil {
		return false, err
	}

	e2 := reflect.New(reflect.TypeOf(m.expected))
	e2.Elem().Set(reflect.ValueOf(m.expected))

	return reflect.DeepEqual(metadata, e2.Interface()), nil
}

func (m *havePersistentMetadataMatcher) FailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf(
		"Expected\n\t%#v\nto have persistent metadata\n\t%#v\n"+
			"but found\n\t%#v\n",
		actual, m.expected, actualMetadata)
}

func (m *havePersistentMetadataMatcher) NegatedFailureMessage(actual interface{}) string {
	actualMetadata, err := m.getMetadata(actual)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf(
		"Expected\n\t%#v\nnot to have persistent metadata\n\t%#v\n"+
			"but found\n\t%#v\n",
		actual, m.expected, actualMetadata)
}

func (m *havePersistentMetadataMatcher) path(actual interface{}) (string, error) {
	v := reflect.ValueOf(actual).FieldByName("Root")
	if v == (reflect.Value{}) {
		return "", fmt.Errorf("HavePersistentMetadata matcher expects a layers")
	}

	return filepath.Join(v.Interface().(string), "store.toml"), nil
}

func (m *havePersistentMetadataMatcher) getMetadata(actual interface{}) (interface{}, error) {
	path, err := m.path(actual)
	if err != nil {
		return nil, err
	}

	in := struct {
		Metadata toml.Primitive `toml:"metadata"`
	}{}

	md, err := toml.DecodeFile(path, &in)
	if err != nil {
		return nil, fmt.Errorf("failed to decode file: %s", err.Error())
	}

	metadata := reflect.New(reflect.TypeOf(m.expected)).Interface()
	if err := md.PrimitiveDecode(in.Metadata, metadata); err != nil {
		return nil, fmt.Errorf("failed to decode: %s", err.Error())
	}

	return metadata, nil
}
