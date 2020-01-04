/*
 * Copyright 2019-2020 the original author or authors.
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

package buildpackplan

import (
	"fmt"

	"github.com/Masterminds/semver"
)

const VersionSource = "version-source"

// PriorityMerge returns a function that merges two Plans together, giving precedence first to the one which declares a
// version; and then, to the plan which has a higher priority level. PriorityMerge defines a Plan's priority level by
// looking for a "version-source" key defined in the Plan.Metadata, and uses the provided map to define the priority
// level for that source. Currently, an unknown source (one specified in a Plan, but not provided in the priority list)
// has a priority level of 0.
// An example priority list, used for the Node Engine CNB, is:
//   "buildpack.yml": 3,
//   "package.json":  2,
//   ".nvmrc":        1,
//   "":              -1
// Metadata for most cases is combined (excluding version-source, build and launch), with a comma delimiter if present
// in both plans. version-source is set to the highest priority between the plans, and build/launch will be set to true
// if either of the plans request them.
func PriorityMerge(priorities map[interface{}]int) MergeFunc {
	return func(a, b Plan) (Plan, error) {
		aVersion := a.Version
		bVersion := b.Version
		aSource := a.Metadata[VersionSource]
		bSource := b.Metadata[VersionSource]

		if aVersion == "" && bVersion == "" {
			return mergePlans(a, b, "", nil)
		} else if aVersion == "" {
			return mergePlans(a, b, bVersion, bSource)
		} else if bVersion == "" {
			return mergePlans(a, b, aVersion, aSource)
		}

		aPriority := getPriority(aSource, priorities)
		bPriority := getPriority(bSource, priorities)
		if aPriority > bPriority {
			return mergePlans(a, b, aVersion, aSource)
		} else if aPriority == bPriority {
			version, err := getHighestVersion(aVersion, bVersion)
			if err != nil {
				return Plan{}, fmt.Errorf("failed to get the highest version between %s and %s: %v", aVersion, bVersion, err)
			}
			return mergePlans(a, b, version, aSource)
		} else {
			return mergePlans(a, b, bVersion, bSource)
		}
	}
}

func getHighestVersion(aVersion, bVersion string) (string, error) {
	aSemver, err := semver.NewVersion(aVersion)
	if err != nil {
		return "", fmt.Errorf("failed to convert version %s to semver", aVersion)
	}
	bSemver, err := semver.NewVersion(bVersion)
	if err != nil {
		return "", fmt.Errorf("failed to convert version %s to semver", bVersion)
	}
	version := aVersion
	if aSemver.LessThan(bSemver) {
		version = bVersion
	}

	return version, nil
}

func getPriority(versionSource interface{}, priorities map[interface{}]int) int {
	val, ok := priorities[versionSource]

	// Any source is higher than empty string
	if !ok {
		val = 0
	}
	return val
}

func mergePlans(a, b Plan, version string, versionSource interface{}) (Plan, error) {
	aBuildVal, err := getBooleanVal(a.Metadata["build"])
	if err != nil {
		return Plan{}, fmt.Errorf("could not determine 'build' metadata of %s: %s", a.Name, err)
	}

	bBuildVal, err := getBooleanVal(b.Metadata["build"])
	if err != nil {
		return Plan{}, fmt.Errorf("could not determine 'build' metadata of %s: %s", b.Name, err)
	}

	aLaunchVal, err := getBooleanVal(a.Metadata["launch"])
	if err != nil {
		return Plan{}, fmt.Errorf("could not determine 'launch' metadata of %s: %s", a.Name, err)
	}

	bLaunchVal, err := getBooleanVal(b.Metadata["launch"])
	if err != nil {
		return Plan{}, fmt.Errorf("could not determine 'launch' metadata of %s: %s", b.Name, err)
	}

	metadata := a.Metadata // NOTE: Mutating metadata also mutates a.Metadata
	for key, val := range b.Metadata {
		ignoreKeys := []string{VersionSource, "build", "launch"}
		if !contains(ignoreKeys, key) && val != "" {
			if aVal, ok := metadata[key]; ok && aVal != "" && aVal != val {
				val = aVal.(string) + "," + val.(string)
			}
			metadata[key] = val
		}
	}

	if versionSource != nil && versionSource != "" {
		metadata[VersionSource] = versionSource
	}

	if aBuildVal || bBuildVal {
		metadata["build"] = true
	}

	if aLaunchVal || bLaunchVal {
		metadata["launch"] = true
	}

	return Plan{
		Name:     a.Name,
		Version:  version,
		Metadata: metadata,
	}, nil
}

func contains(slice []string, val string) bool {
	for _, x := range slice {
		if x == val {
			return true
		}
	}

	return false
}

func getBooleanVal(val interface{}) (bool, error) {
	if val == nil || val == "" {
		return false, nil
	}

	if b, isString := val.(string); isString {
		return b == "true", nil
	} else if b, isBool := val.(bool); isBool {
		return b, nil
	}

	return false, fmt.Errorf("could not get boolean value of %v", val)
}
