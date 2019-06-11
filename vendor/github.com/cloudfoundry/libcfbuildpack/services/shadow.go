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

package services

import (
	"github.com/buildpack/libbuildpack/services"
)

// Credentials is the collection of credential keys.
//
// In order to encourage good design, this does not include the values even though they exist.  Buildpacks should
// only extract values at startup/runtime and not embed them in image.
type Credentials = services.Credentials

// Service represents a service bound to the application.
type Service = services.Service
