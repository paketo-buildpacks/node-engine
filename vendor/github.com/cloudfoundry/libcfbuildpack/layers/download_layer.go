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

package layers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/fatih/color"
)

// DownloadLayer is an extension to Layer that is unique to a dependency download.
type DownloadLayer struct {
	Layer

	cacheLayer Layer
	dependency buildpack.Dependency
	info       buildpack.Info
	logger     logger.Logger
}

// Artifact returns the path to an artifact cached in the layer.  If the artifact has already been downloaded, the cache
// will be validated and used directly.  If the artifact is out of date, the layer is left untouched and the contributor
// is responsible for cleaning the layer if necessary.
func (l DownloadLayer) Artifact() (string, error) {
	l.Touch()

	matches, err := l.cacheLayer.MetadataMatches(l.dependency)
	if err != nil {
		return "", err
	}

	artifact := filepath.Join(l.cacheLayer.Root, filepath.Base(l.dependency.URI))
	if matches {
		l.logger.SubsequentLine("%s cached download from buildpack", color.GreenString("Reusing"))
		return artifact, nil
	}

	matches, err = l.MetadataMatches(l.dependency)
	if err != nil {
		return "", err
	}

	artifact = filepath.Join(l.Root, filepath.Base(l.dependency.URI))
	if matches {
		l.logger.SubsequentLine("%s cached download from previous build", color.GreenString("Reusing"))
		return artifact, nil
	}

	if err := os.RemoveAll(l.Root); err != nil {
		return "", err
	}

	l.logger.SubsequentLine("%s from %s", color.YellowString("Downloading"), l.dependency.URI)
	if err := l.download(artifact); err != nil {
		return "", err
	}

	l.logger.SubsequentLine("Verifying checksum")
	if err := l.verify(artifact); err != nil {
		return "", err
	}

	if err := l.WriteMetadata(l.dependency, Cache); err != nil {
		return "", err
	}

	return artifact, nil
}

func (l DownloadLayer) download(file string) error {
	req, err := http.NewRequest("GET", l.dependency.URI, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", l.info.ID, l.info.Version))
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	client := http.Client{Transport: t}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("could not download: %bd", resp.StatusCode)
	}

	return helper.WriteFileFromReader(file, 0644, resp.Body)
}

func (l DownloadLayer) verify(file string) error {
	s := sha256.New()

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(s, f)
	if err != nil {
		return err
	}

	actualSha256 := hex.EncodeToString(s.Sum(nil))

	if actualSha256 != l.dependency.SHA256 {
		return fmt.Errorf("dependency sha256 mismatch: expected sha256 %s, actual sha256 %s",
			l.dependency.SHA256, actualSha256)
	}
	return nil
}
