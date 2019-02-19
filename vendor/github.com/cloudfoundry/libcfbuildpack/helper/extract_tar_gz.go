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

package helper

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

// ExtractTarGz extracts source GZIP'd TAR file to a destination directory.  An arbitrary number of top-level directory
// components can be stripped from each path.
func ExtractTarGz(source string, destination string, stripComponents int) error {
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	return extractTar(gz, destination, stripComponents)
}

func extractTar(source io.Reader, destination string, stripComponents int) error {
	t := tar.NewReader(source)

	for {
		f, err := t.Next()
		if err == io.EOF {
			break
		}

		target := strippedPath(f.Name, destination, stripComponents)
		if target == "" {
			continue
		}

		info := f.FileInfo()
		if info.IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		} else if info.Mode()&os.ModeSymlink != 0 {
			if err := WriteSymlink(f.Linkname, target); err != nil {
				return err
			}
		} else {
			if err := WriteFileFromReader(target, info.Mode(), t); err != nil {
				return err
			}
		}
	}

	return nil
}
