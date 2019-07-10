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
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libcfbuildpack/internal"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/fatih/color"
)

// TouchedLayers contains information about the layers that have been touched as part of this execution.
type TouchedLayers struct {
	// Root is the root location of all layers to inspect for unused layers.
	Root string

	logger  logger.Logger
	touched internal.Set
}

// Add registers that a given layer has been touched
func (t TouchedLayers) Add(metadata string) {
	t.logger.Debug("Layer %s touched", metadata)
	t.touched.Add(metadata)
}

// Cleanup removes all layers that have not been touched as part of this execution.
func (t TouchedLayers) Cleanup() error {
	candidates, err := t.candidates()
	if err != nil {
		return err
	}

	if t.logger.IsDebugEnabled() {
		t.logger.Debug("Existing Layers: %s", candidates)
		t.logger.Debug("Touched Layers: %s", t.touched)
	}

	remove := candidates.Difference(t.touched)
	if remove.Size() == 0 {
		return nil
	}

	t.logger.Header("%s unused layers", color.YellowString("Removing"))
	for r := range remove.Iterator() {
		f := r.(string)
		t.logger.Body(strings.TrimSuffix(filepath.Base(f), ".toml"))

		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}

	return nil
}

func (t TouchedLayers) candidates() (internal.Set, error) {
	files, err := filepath.Glob(filepath.Join(t.Root, "*.toml"))
	if err != nil {
		return internal.Set{}, err
	}

	candidates := internal.NewSet()

	launch := filepath.Join(t.Root, "launch.toml")
	store := filepath.Join(t.Root, "store.toml")
	for _, f := range files {
		if f != launch && f != store {
			candidates.Add(f)
		}
	}

	return candidates, nil
}

// NewTouchedLayers creates a new instance that monitors a given root.
func NewTouchedLayers(root string, logger logger.Logger) TouchedLayers {
	return TouchedLayers{root, logger, internal.NewSet()}
}
