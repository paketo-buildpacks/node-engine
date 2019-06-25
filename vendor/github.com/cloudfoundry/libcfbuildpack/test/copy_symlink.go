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
	"os"
	"path/filepath"
	"testing"
)

// CopySymlink copies source to destination.  Before writing, it creates all required parent directories for the
// destination.
func CopySymlink(t *testing.T, source string, destination string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		t.Fatal(err)
	}

	target, err := os.Readlink(source)
	if err != nil {
		t.Fatal(fmt.Errorf("error while reading symlink '%s': %v", source, err))
	}

	WriteSymlink(t, target, destination)
}
