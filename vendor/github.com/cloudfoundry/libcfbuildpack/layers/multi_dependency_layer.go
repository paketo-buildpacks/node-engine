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
	"os"
	"strings"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

type MultiDependencyLayer struct {
	Layer

	// Dependencies are the dependencies provided by this layer.
	Dependencies []buildpack.Dependency

	dependencyBuildPlans buildplan.BuildPlan
	downloadLayers       []DownloadLayer
	logger               logger.Logger
}

// MultiDependencyLayerContributor defines a callback function that is called when a dependency needs to be contributed.
type MultiDependencyLayerContributor func(artifact string, layer MultiDependencyLayer) error

// Contribute facilitates custom contribution of an artifacts to a layer.  If the artifacts have already been
// contributed, the contribution is validated and the contributor is not called.  If the contribution is out of date,
// the layer is completely removed before contribution occurs.
func (l MultiDependencyLayer) Contribute(contributors map[string]MultiDependencyLayerContributor, flags ...Flag) error {
	for _, dl := range l.downloadLayers {
		dl.Touch()
	}

	if err := l.Layer.Contribute(multiDependency{l.Dependencies}, func(layer Layer) error {
		if err := os.RemoveAll(l.Root); err != nil {
			return err
		}

		for i, d := range l.Dependencies {
			c, ok := contributors[d.ID]
			if !ok {
				return fmt.Errorf("no contributor found for dependency %s", d.ID)
			}

			a, err := l.downloadLayers[i].Artifact()
			if err != nil {
				return err
			}

			if err := c(a, l); err != nil {
				return err
			}
		}

		return nil
	}, flags...); err != nil {
		return err;
	}

	l.contributeToBuildPlan()
	return nil
}

func (l *MultiDependencyLayer) contributeToBuildPlan() {
	for _, d := range l.Dependencies {
		l.Logger.Debug("Contributing %s to bill-of-materials", d.ID)

		l.dependencyBuildPlans[d.ID] = buildplan.Dependency{
			Version: d.Version.Original(),
			Metadata: buildplan.Metadata{
				"name":     d.Name,
				"uri":      d.URI,
				"sha256":   d.SHA256,
				"stacks":   d.Stacks,
				"licenses": d.Licenses,
			},
		}
	}
}

type multiDependency struct {
	Dependencies []buildpack.Dependency `toml:"dependencies"`
}

func (m multiDependency) Identity() (string, string) {
	var names []string
	for _, d := range m.Dependencies {
		names = append(names, d.Name)
	}

	if m.Dependencies[0].Version.Version != nil {
		return strings.Join(names, ", "), m.Dependencies[0].Version.Version.Original()
	}

	return strings.Join(names, ", "), ""
}
