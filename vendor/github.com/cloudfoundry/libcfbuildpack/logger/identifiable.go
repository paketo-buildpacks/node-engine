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

package logger

import (
	"strings"
)

// Identifiable is an interface that indicates that a type has an identity.
type Identifiable interface {
	// Identity is the method that returns the required name and optional description that make up identity.
	Identity() (name string, description string)
}

// PrettyIdentity formats a standard pretty identity of a type.
func PrettyIdentity(v Identifiable) string {
	if v == nil {
		return ""
	}

	var sb strings.Builder

	name, description := v.Identity()

	_, _ = sb.WriteString(name)

	if description != "" {
		_, _ = sb.WriteString(" ")
		_, _ = sb.WriteString(description)
	}

	return sb.String()
}
