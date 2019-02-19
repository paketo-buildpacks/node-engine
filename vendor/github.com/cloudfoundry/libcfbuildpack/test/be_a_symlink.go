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

// BeASymlink asserts that a file is a symlink and the link points to a given target.
func BeASymlink(target string) types.GomegaMatcher {
	return &beASymlinkMatcher{
		target: target,
	}
}

type beASymlinkMatcher struct {
	target string
}

func (m *beASymlinkMatcher) Match(actual interface{}) (bool, error) {
	path, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("BeASymlink matcher expects a path")
	}

	fi, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("Unable to stat file: %s", err.Error())
	}

	if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
		return false, nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return false, fmt.Errorf("Unable to read link :%s", err.Error())
	}

	return target == m.target, nil
}

func (m *beASymlinkMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be a symlink to\n\t%#v", actual, m.target)
}

func (m *beASymlinkMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be a symlink to\n\t%#v", actual, m.target)
}
