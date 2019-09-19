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
	"fmt"
	"sort"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/fatih/color"
)

// Layers is an extension allows additional functionality to be added.
type Layers struct {
	layers.Layers

	// Plans contains all contributed dependencies.
	Plans *buildpackplan.Plans

	// TouchedLayers registers the layers that have been touched during this execution.
	TouchedLayers TouchedLayers

	buildpack      buildpack.Buildpack
	buildpackCache layers.Layers
	logger         logger.Logger
}

// DependencyLayer returns a DependencyLayer unique to a dependency.
func (l Layers) DependencyLayer(dependency buildpack.Dependency) DependencyLayer {
	return l.DependencyLayerWithID(dependency.ID, dependency)
}

// DependencyLayerWithID returns a DependencyLayer unique to a dependency with an explicit id.
func (l Layers) DependencyLayerWithID(id string, dependency buildpack.Dependency) DependencyLayer {
	return DependencyLayer{
		l.Layer(id),
		dependency,
		l.DownloadLayer(dependency),
		l.logger,
		l.Plans,
	}
}

// DownloadLayer returns a DownloadLayer unique to a dependency.
func (l Layers) DownloadLayer(dependency buildpack.Dependency) DownloadLayer {
	return DownloadLayer{
		l.Layer(dependency.SHA256),
		Layer{l.buildpackCache.Layer(dependency.SHA256), l.logger, l.TouchedLayers},
		dependency,
		l.buildpack.Info,
		l.logger,
	}
}

// HelperLayer returns a HelperLayer unique to a buildpack provided dependency.
func (l Layers) HelperLayer(id string, name string) HelperLayer {
	return HelperLayer{
		l.Layer(id),
		id,
		l.buildpack,
		l.logger,
		name,
		l.Plans,
	}
}

// Layer creates a Layer with a specified name.
func (l Layers) Layer(name string) Layer {
	return Layer{l.Layers.Layer(name), l.logger, l.TouchedLayers}
}

// WriteApplicationMetadata writes application metadata to the filesystem.
func (l Layers) WriteApplicationMetadata(metadata Metadata) error {
	if len(metadata.Slices) > 0 {
		l.logger.Header("%d application slices", len(metadata.Slices))
	}

	if len(metadata.Processes) > 0 {
		l.logger.Header("Process types:")

		p := metadata.Processes
		sort.Slice(p, func(i int, j int) bool {
			return p[i].Type < p[j].Type
		})

		max := l.maximumTypeLength(p)
		for _, p := range p {
			format := fmt.Sprintf("%%s%%s:%%-%ds %%s", max-len(p.Type))
			l.logger.Info(format, logger.BodyIndent, color.CyanString(p.Type), "", p.Command)
		}
	}

	return l.Layers.WriteApplicationMetadata(metadata)
}

// WritePersistentMetadata writes persistent metadata to the filesystem.
func (l Layers) WritePersistentMetadata(metadata interface{}) error {
	l.logger.Body("Writing persistent metadata")
	return l.Layers.WritePersistentMetadata(metadata)
}

func (l Layers) maximumTypeLength(processes Processes) int {
	max := 0

	for _, t := range processes {
		if l := len(t.Type); l > max {
			max = l
		}
	}

	return max
}

// NewLayers creates a new instance of Layers.
func NewLayers(layers layers.Layers, buildpackCache layers.Layers, buildpack buildpack.Buildpack, logger logger.Logger) Layers {
	return Layers{
		Layers:         layers,
		Plans:          &buildpackplan.Plans{},
		TouchedLayers:  NewTouchedLayers(layers.Root, logger),
		buildpack:      buildpack,
		buildpackCache: buildpackCache,
		logger:         logger,
	}
}
