/*
 * Copyright 2018-2020 the original author or authors.
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

	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// MultiDependencyLayer is an extension to Layer that is unique to a collection of dependencies.
type MultiDependencyLayer struct {
	Layer

	// Dependencies are the dependencies provided by this layer.
	Dependencies []buildpack.Dependency

	downloadLayers map[string]DownloadLayer
	logger         logger.Logger
	plans          *buildpackplan.Plans
}

// MultiDependencyLayerContributor defines a callback function that is called when a dependency needs to be contributed.
type MultiDependencyLayerContributor func(artifact string, layer MultiDependencyLayer) error

func (l MultiDependencyLayer) Contribute(contributors map[string]MultiDependencyLayerContributor, flags ...Flag) error {
	for _, v := range l.downloadLayers {
		v.Touch()
	}

	if err := l.Layer.Contribute(metadata(l.Dependencies), func(layer Layer) error {
		if err := os.RemoveAll(l.Root); err != nil {
			return err
		}

		for _, d := range l.Dependencies {
			dl, ok := l.downloadLayers[d.ID]
			if !ok {
				return fmt.Errorf("unable to find download layer for %s", d.ID)
			}

			a, err := dl.Artifact()
			if err != nil {
				return err
			}

			c, ok := contributors[d.ID]
			if !ok {
				return fmt.Errorf("unable to find contributor for %s", d.ID)
			}

			if err := c(a, l); err != nil {
				return err
			}
		}

		return nil
	}, flags...); err != nil {
		return err
	}

	l.contributeToBuildPlan()
	return nil
}

func (l *MultiDependencyLayer) contributeToBuildPlan() {
	for _, d := range l.Dependencies {
		l.logger.Debug("Contributing %s to bill-of-materials", d.ID)

		l.plans.Entries = append(l.plans.Entries, buildpackplan.Plan{
			Name:    d.ID,
			Version: d.Version.Original(),
			Metadata: buildpackplan.Metadata{
				"name":     d.Name,
				"uri":      d.URI,
				"sha256":   d.SHA256,
				"stacks":   d.Stacks,
				"licenses": d.Licenses,
			},
		})
	}
}

type metadata []buildpack.Dependency

func (m metadata) Identity() (string, string) {
	if len(m) > 0 {
		return m[0].Identity()
	}

	return "", ""
}
