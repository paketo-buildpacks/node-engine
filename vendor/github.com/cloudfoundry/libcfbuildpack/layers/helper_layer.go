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

package layers

import (
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/buildpack"
	"github.com/cloudfoundry/libcfbuildpack/logger"
)

// HelperLayer is an extension to Layer that is unique to a buildpack provided helper.
type HelperLayer struct {
	Layer

	// ID is the id of the buildpack provided helper.
	ID string

	buildpack            buildpack.Buildpack
	dependencyBuildPlans buildplan.BuildPlan
	name                 string
	logger               logger.Logger
}

// HelperLayerContributor defines a callback function that is called when a buildpack provided helper needs to be
// contributed.
type HelperLayerContributor func(artifact string, layer HelperLayer) error

// Contribute facilitates custom contribution of a buildpack provided helper to a layer.  If the artifact has already
// been contributed, the contribution is validated and the contributor is not called.  If the contribution is out of
// date, the layer is completely removed before contribution occurs.
func (l HelperLayer) Contribute(contributor HelperLayerContributor, flags ...Flag) error {
	if err := l.Layer.Contribute(marker{l.buildpack.Info, l.name}, func(layer Layer) error {
		if err := os.RemoveAll(l.Root); err != nil {
			return err
		}

		a := filepath.Join(l.buildpack.Root, "bin", l.ID)

		return contributor(a, l)
	}, flags...); err != nil {
		return err
	}

	l.contributeToBuildPlan()
	return nil
}

func (l *HelperLayer) contributeToBuildPlan() {
	l.logger.Debug("Contributing %s to bill-of-materials", l.ID)

	l.dependencyBuildPlans[l.ID] = buildplan.Dependency{
		Version: l.buildpack.Info.Version,
		Metadata: buildplan.Metadata{
			"id":   l.buildpack.Info.ID,
			"name": l.buildpack.Info.Name,
		},
	}
}

type marker struct {
	buildpack.Info

	DisplayName string `toml:"display_name"`
}

func (m marker) Identity() (string, string) {
	return m.DisplayName, m.Version
}
