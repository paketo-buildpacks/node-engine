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

package test

import (
	"fmt"
	"os"

	"github.com/onsi/gomega/types"
)

func HavePermissions(expected os.FileMode) types.GomegaMatcher {
	return &havePermissionsMatcher{
		expected: expected,
	}
}

type havePermissionsMatcher struct {
	expected os.FileMode
}

func (m *havePermissionsMatcher) Match(actual interface{}) (bool, error) {
	filename, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("HaveContent matcher expects an filename")
	}

	fi, err := os.Stat(filename)
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %s", err.Error())
	}

	return fi.Mode() == m.expected, nil
}

func (m *havePermissionsMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto have permissions of of\n\t%#v", actual, m.expected.String())
}

func (m *havePermissionsMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to have permissions of\n\t%#v", actual, m.expected.String())
}
