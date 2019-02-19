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

package logger

import (
	"fmt"
	"strings"

	"github.com/buildpack/libbuildpack/logger"
	"github.com/fatih/color"
)

const indent = "      "

var eyeCatcher string

func init() {
	color.NoColor = false
	eyeCatcher = color.New(color.FgRed, color.Bold).Sprint("----->")
}

// Logger is an extension to libbuildpack.Logger to add additional functionality.
type Logger struct {
	logger.Logger
}

// FirstLine prints the log messages with the first line eye catcher.
func (l Logger) FirstLine(format string, args ...interface{}) {
	if !l.IsInfoEnabled() {
		return
	}

	l.Info("%s %s", eyeCatcher, fmt.Sprintf(format, args...))
}

// SubsequentLine prints log message with the subsequent line indent.
func (l Logger) SubsequentLine(format string, args ...interface{}) {
	if !l.IsInfoEnabled() {
		return
	}

	l.Info("%s %s", indent, fmt.Sprintf(format, args...))
}

// PrettyIdentity formats a standard pretty identity of a type.
func (l Logger) PrettyIdentity(v Identifiable) string {
	if v == nil {
		return ""
	}

	var sb strings.Builder

	name, description := v.Identity()

	_, _ = sb.WriteString(color.New(color.FgBlue, color.Bold).Sprint(name))

	if description != "" {
		_, _ = sb.WriteString(" ")
		_, _ = sb.WriteString(color.BlueString(description))
	}

	return sb.String()
}

// String makes Logger satisfy the Stringer interface.
func (l Logger) String() string {
	return fmt.Sprintf("Logger{ Logger: %s }", l.Logger)
}
