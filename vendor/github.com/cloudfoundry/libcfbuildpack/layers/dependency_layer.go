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

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// DependencyLayer is an extension to Layer that is unique to a dependency.
type DependencyLayer struct {
	Layer

	// Dependency is the dependency provided by this layer.
	Dependency buildpack.Dependency

	dependencyBuildPlans buildplan.BuildPlan
	downloadLayer        DownloadLayer
	logger               logger.Logger
}

// ArtifactName returns the name portion of the download path for the dependency.
func (l DependencyLayer) ArtifactName() string {
	return filepath.Base(l.Dependency.URI)
}

// DependencyLayerContributor defines a callback function that is called when a dependency needs to be contributed.
type DependencyLayerContributor func(artifact string, layer DependencyLayer) error

// Contribute facilitates custom contribution of an artifact to a layer.  If the artifact has already been contributed,
// the contribution is validated and the contributor is not called.  If the contribution is out of date, the layer is
// completely removed before contribution occurs.
func (l DependencyLayer) Contribute(contributor DependencyLayerContributor, flags ...Flag) error {
	l.downloadLayer.Touch()

	if err := l.Layer.Contribute(l.Dependency, func(layer Layer) error {
		if err := os.RemoveAll(l.Root); err != nil {
			return err
		}

		a, err := l.downloadLayer.Artifact()
		if err != nil {
			return err
		}

		return contributor(a, l)
	}, flags...); err != nil {
		return err
	}

	l.contributeToBuildPlan()
	return nil
}

func (l *DependencyLayer) contributeToBuildPlan() {
	l.logger.Debug("Contributing %s to bill-of-materials", l.Dependency.ID)

	l.dependencyBuildPlans[l.Dependency.ID] = buildplan.Dependency{
		Version: l.Dependency.Version.Original(),
		Metadata: buildplan.Metadata{
			"name":     l.Dependency.Name,
			"uri":      l.Dependency.URI,
			"sha256":   l.Dependency.SHA256,
			"stacks":   l.Dependency.Stacks,
			"licenses": l.Dependency.Licenses,
		},
	}
}
