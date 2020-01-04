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

package services

import (
	"strings"

	"github.com/buildpack/libbuildpack/services"
)

// Services is a collection of services bound to the application.
type Services struct {
	services.Services
}

// FindServiceCredentials returns the credentials payload for given service.  The selected service is one who's
// BindingName, InstanceName, Label, or Tags contain the filter and has the required credentials.  Returns the
// credentials and true if exactly one service is matched, otherwise false.
//
// NOTE: This function should ONLY be used to extract values that are CONSTANT throughout the lifecycle of a staged
// application.  Typically this means you should only use this function to get agent download information from Service
// Brokers and SHOULD NOT EVER retrieve and use connection credentials.
func (s Services) FindServiceCredentials(filter string, credentials ...string) (Credentials, bool) {
	match := make([]Service, 0)

	for _, c := range s.Services {
		if s.matchesService(c, filter) && s.matchesCredentials(c, credentials) {
			match = append(match, c)
		}
	}

	if len(match) != 1 {
		return nil, false
	}

	return match[0].Credentials, true
}

// HasService determines whether a single service, who's BindingName, InstanceName, Label, or Tags contain the filter
// and has the required credentials, exists.  Returns true if exactly one service is matched, false otherwise.
func (s Services) HasService(filter string, credentials ...string) bool {
	var match []Service

	for _, c := range s.Services {
		if s.matchesService(c, filter) && s.matchesCredentials(c, credentials) {
			match = append(match, c)
		}
	}

	return len(match) == 1
}

type biPredicate func(x string, y string) bool

func (Services) any(test biPredicate, s string, candidates []string) bool {
	for _, c := range candidates {
		if test(c, s) {
			return true
		}
	}

	return false
}

func (Services) equality(x string, y string) bool {
	return x == y
}

func (Services) matchesBindingName(service Service, filter string) bool {
	return strings.Contains(service.BindingName, filter)
}

func (s Services) matchesCredentials(service Service, credentials []string) bool {
	cr := service.Credentials

	candidates := make([]string, 0, len(cr))
	for k := range cr {
		candidates = append(candidates, k)
	}

	for _, c := range credentials {
		if !s.any(s.equality, c, candidates) {
			return false
		}
	}

	return true
}

func (Services) matchesInstanceName(service Service, filter string) bool {
	return strings.Contains(service.InstanceName, filter)
}

func (Services) matchesLabel(service Service, filter string) bool {
	return strings.Contains(service.Label, filter)
}

func (s Services) matchesService(service Service, filter string) bool {
	return s.matchesBindingName(service, filter) ||
		s.matchesInstanceName(service, filter) ||
		s.matchesLabel(service, filter) ||
		s.matchesTag(service, filter)
}

func (s Services) matchesTag(service Service, filter string) bool {
	return s.any(strings.Contains, filter, service.Tags)
}
