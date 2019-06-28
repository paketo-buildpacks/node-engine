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
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/buildpack/libbuildpack/internal"
	"github.com/buildpack/libbuildpack/logger"
)

// Layer represents a layer for an application.
type Layer struct {
	// Root is the path to the root directory for the layer.
	Root string

	// Metadata is the location of the layer's metadata file.
	Metadata string

	logger logger.Logger
}

// AppendBuildEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendBuildEnv(name string, format string, args ...interface{}) error {
	return l.addBuildEnvFile(fmt.Sprintf("%s.append", name), format, args...)
}

// AppendLaunchEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendLaunchEnv(name string, format string, args ...interface{}) error {
	return l.addLaunchEnvFile(fmt.Sprintf("%s.append", name), format, args...)
}

// AppendSharedEnv appends the value of this environment variable to any previous declarations of the value without any
// delimitation.  If delimitation is important during concatenation, callers are required to add it.
func (l Layer) AppendSharedEnv(name string, format string, args ...interface{}) error {
	return l.addSharedEnvFile(fmt.Sprintf("%s.append", name), format, args...)
}

// AppendPathBuildEnv appends the value of this environment variable to any previous declarations of the value using the
// OS path delimiter.
func (l Layer) AppendPathBuildEnv(name string, format string, args ...interface{}) error {
	return l.addBuildEnvFile(name, format, args...)
}

// AppendPathLaunchEnv appends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
func (l Layer) AppendPathLaunchEnv(name string, format string, args ...interface{}) error {
	return l.addLaunchEnvFile(name, format, args...)
}

// AppendPathSharedEnv appends the value of this environment variable to any previous declarations of the value using
// the OS path delimiter.
func (l Layer) AppendPathSharedEnv(name string, format string, args ...interface{}) error {
	return l.addSharedEnvFile(name, format, args...)
}

// OverrideBuildEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideBuildEnv(name string, format string, args ...interface{}) error {
	return l.addBuildEnvFile(fmt.Sprintf("%s.override", name), format, args...)
}

// OverrideLaunchEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideLaunchEnv(name string, format string, args ...interface{}) error {
	return l.addLaunchEnvFile(fmt.Sprintf("%s.override", name), format, args...)
}

// OverrideSharedEnv overrides any existing value for an environment variable with this value.
func (l Layer) OverrideSharedEnv(name string, format string, args ...interface{}) error {
	return l.addSharedEnvFile(fmt.Sprintf("%s.override", name), format, args...)
}

// ReadMetadata reads arbitrary layer metadata from the filesystem.
func (l Layer) ReadMetadata(metadata interface{}) error {
	exists, err := internal.FileExists(l.Metadata)
	if err != nil {
		return err
	}

	if !exists {
		l.logger.Debug("Metadata %s does not exist", l.Metadata)
		return nil
	}

	in := struct {
		Metadata toml.Primitive `toml:"metadata"`
	}{}

	md, err := toml.DecodeFile(l.Metadata, &in)
	if err != nil {
		return err
	}

	if err := md.PrimitiveDecode(in.Metadata, metadata); err != nil {
		return err
	}

	l.logger.Debug("Reading layer metadata: %s => %v", l.Metadata, metadata)
	return nil
}

// RemoveMetadata remove layer metadata from the filesystem.
func (l Layer) RemoveMetadata() error {
	exists, err := internal.FileExists(l.Metadata)
	if err != nil {
		return err
	}

	if !exists {
		l.logger.Debug("Metadata %s does not exist", l.Metadata)
		return nil
	}

	return os.Remove(l.Metadata)
}

// WriteMetadata writes arbitrary layer metadata to the filesystem.
func (l Layer) WriteMetadata(metadata interface{}, flags ...Flag) error {
	lm := layerMetadata{Metadata: metadata}

	for _, flag := range flags {
		switch flag {
		case Build:
			lm.Build = true
		case Cache:
			lm.Cache = true
		case Launch:
			lm.Launch = true
		}
	}

	l.logger.Debug("Writing layer metadata: %s <= %#v", l.Metadata, lm)
	return internal.WriteTomlFile(l.Metadata, 0644, lm)
}

// WriteProfile writes a file to profile.d with this value.
func (l Layer) WriteProfile(file string, format string, args ...interface{}) error {
	f := filepath.Join(l.Root, "profile.d", file)

	if l.logger.IsDebugEnabled() {
		l.logger.Debug("Writing profile: %s <= %s", f, fmt.Sprintf(format, args...))
	}

	return internal.WriteFile(f, 0644, format, args...)
}

func (l Layer) addBuildEnvFile(file string, format string, args ...interface{}) error {
	return l.addEnvFile(filepath.Join("env.build", file), format, args...)
}

func (l Layer) addEnvFile(file string, format string, args ...interface{}) error {
	f := filepath.Join(l.Root, file)

	if l.logger.IsDebugEnabled() {
		l.logger.Debug("Writing environment variable: %s <= %s", f, fmt.Sprintf(format, args...))
	}

	return internal.WriteFile(f, 0644, format, args...)
}

func (l Layer) addLaunchEnvFile(file string, format string, args ...interface{}) error {
	return l.addEnvFile(filepath.Join("env.launch", file), format, args...)
}

func (l Layer) addSharedEnvFile(file string, format string, args ...interface{}) error {
	return l.addEnvFile(filepath.Join("env", file), format, args...)
}

type layerMetadata struct {
	Build    bool        `toml:"build"`
	Cache    bool        `toml:"cache"`
	Launch   bool        `toml:"launch"`
	Metadata interface{} `toml:"metadata"`
}
