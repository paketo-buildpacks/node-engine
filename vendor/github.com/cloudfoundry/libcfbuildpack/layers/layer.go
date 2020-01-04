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

package layers

import (
	"reflect"

	"github.com/buildpack/libbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/logger"
	"github.com/fatih/color"
)

var identityColor = color.New(color.FgBlue)

// LaunchLayer is an extension to libbuildpack.LaunchLayer that allows additional functionality to be added
type Layer struct {
	layers.Layer

	// Logger is used to write debug and info to the console.
	Logger logger.Logger

	touchedLayers TouchedLayers
}

// AppendBuildEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendBuildEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.AppendBuildEnv(name, format, args...)
}

// AppendLaunchEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendLaunchEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to launch", name)
	return l.Layer.AppendLaunchEnv(name, format, args...)
}

// AppendSharedEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendSharedEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.AppendSharedEnv(name, format, args...)
}

// AppendPathBuildEnv appends the value of this environment variable to any previous declarations of the value using the
// OS path delimiter.
//
// Deprecated: Use PrependPathBuildEnv
func (l Layer) AppendPathBuildEnv(name string, format string, args ...interface{}) error {
	return l.PrependPathBuildEnv(name, format, args...)
}

// AppendPathLaunchEnv appends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
//
// Deprecated: Use PrependPathLaunchEnv
func (l Layer) AppendPathLaunchEnv(name string, format string, args ...interface{}) error {
	return l.PrependPathLaunchEnv(name, format, args...)
}

// AppendPathSharedEnv appends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
//
// Deprecated: Use PrependPathSharedEnv
func (l Layer) AppendPathSharedEnv(name string, format string, args ...interface{}) error {
	return l.PrependPathSharedEnv(name, format, args...)
}

// DefaultBuildEnv sets a default for an environment variable with this value.
func (l Layer) DefaultBuildEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.DefaultBuildEnv(name, format, args...)
}

// DefaultLaunchEnv sets a default for an environment variable with this value.
func (l Layer) DefaultLaunchEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to launch", name)
	return l.Layer.DefaultLaunchEnv(name, format, args...)
}

// DefaultSharedEnv sets a default for an environment variable with this value.
func (l Layer) DefaultSharedEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.DefaultSharedEnv(name, format, args...)
}

// DelimiterBuildEnv sets a delimiter for an environment variable with this value.
func (l Layer) DelimiterBuildEnv(name string, delimiter string) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.DelimiterBuildEnv(name, delimiter)
}

// DelimiterLaunchEnv sets a delimiter for an environment variable with this value.
func (l Layer) DelimiterLaunchEnv(name string, delimiter string) error {
	l.Touch()
	l.Logger.Body("Writing %s to launch", name)
	return l.Layer.DelimiterLaunchEnv(name, delimiter)
}

// DelimiterSharedEnv sets a delimiter for an environment variable with this value.
func (l Layer) DelimiterSharedEnv(name string, delimiter string) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.DefaultSharedEnv(name, delimiter)
}

// OverrideBuildEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideBuildEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.OverrideBuildEnv(name, format, args...)
}

// OverrideLaunchEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideLaunchEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to launch", name)
	return l.Layer.OverrideLaunchEnv(name, format, args...)
}

// OverrideSharedEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideSharedEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.OverrideSharedEnv(name, format, args...)
}

// PrependBuildEnv prepends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) PrependBuildEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.PrependBuildEnv(name, format, args...)
}

// PrependLaunchEnv prepends the value of this environment variable to any previous declarations of the value without
// any delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) PrependLaunchEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.PrependSharedEnv(name, format, args...)
}

// PrependSharedEnv prepends the value of this environment variable to any previous declarations of the value without
// any delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) PrependSharedEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.PrependSharedEnv(name, format, args...)
}

// PrependPathBuildEnv prepends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
func (l Layer) PrependPathBuildEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to build", name)
	return l.Layer.PrependPathBuildEnv(name, format, args...)
}

// PrependPathLaunchEnv prepends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
func (l Layer) PrependPathLaunchEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to launch", name)
	return l.Layer.PrependPathLaunchEnv(name, format, args...)
}

// PrependPathSharedEnv prepends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
func (l Layer) PrependPathSharedEnv(name string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing %s to shared", name)
	return l.Layer.PrependPathSharedEnv(name, format, args...)
}

// LayerContributor defines a callback function that is called when a layer needs to be contributed.
type LayerContributor func(layer Layer) error

// Contribute facilitates custom contribution of a layer.  If the layer has already been contributed, the contribution
// is validated and the contributor is not called.  If the contribution is out of date, the layer is
// // completely removed before contribution occurs.
func (l Layer) Contribute(expected logger.Identifiable, contributor LayerContributor, flags ...Flag) error {
	l.Touch()

	matches, err := l.MetadataMatches(expected)
	if err != nil {
		return err
	}

	if matches {
		l.Logger.Header("%s: %s cached layer",
			l.prettyIdentity(expected), color.GreenString("Reusing"))
		return l.WriteMetadata(expected, flags...)
	}

	l.Logger.Header("%s: %s to layer",
		l.prettyIdentity(expected), color.YellowString("Contributing"))

	if err := contributor(l); err != nil {
		l.Logger.Debug("Error during contribution")
		return err
	}

	return l.WriteMetadata(expected, flags...)
}

// MetadataMatches compares the expected metadata for the actual metadata of this layer.
func (l Layer) MetadataMatches(expected interface{}) (bool, error) {
	l.Touch()

	if expected == nil {
		return false, nil
	}

	actual := reflect.New(reflect.TypeOf(expected)).Interface()

	if err := l.ReadMetadata(actual); err != nil {
		l.Logger.Debug("Dependency metadata is not structured correctly: %s", err.Error())
		return false, nil
	}

	e2 := reflect.New(reflect.TypeOf(expected))
	e2.Elem().Set(reflect.ValueOf(expected))

	matches := reflect.DeepEqual(actual, e2.Interface())
	if !matches {
		l.Logger.Debug("Layer metadata %s does not match expected %s", actual, expected)
	}

	return matches, nil
}

// Touch touches a layer, indicating that it was used and should not be removed.
func (l Layer) Touch() {
	l.touchedLayers.Add(l.Metadata)
}

// WriteProfile writes a file to profile.d with this value.
func (l Layer) WriteProfile(file string, format string, args ...interface{}) error {
	l.Touch()
	l.Logger.Body("Writing .profile.d/%s", file)
	return l.Layer.WriteProfile(file, format, args...)
}

func (Layer) prettyIdentity(v logger.Identifiable) string {
	if v == nil {
		return ""
	}

	name, description := v.Identity()

	if description == "" {
		return identityColor.Sprint(name)
	}

	return identityColor.Sprintf("%s %s", name, description)
}
