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

package services

import (
	"encoding/json"
	"os"

	"github.com/buildpack/libbuildpack/logger"
	"github.com/buildpack/libbuildpack/platform"
)

// Services is a collection of services bound to the application.
type Services []Service

// DefaultServices creates a new instance of Services.
func DefaultServices(platform platform.Platform, logger logger.Logger) (Services, error) {
	s, ok := os.LookupEnv("CNB_SERVICES")

	if !ok {
		s, ok = platform.EnvironmentVariables["CNB_SERVICES"]
	}

	if !ok {
		return Services{}, nil
	}

	var in map[string][]json.RawMessage
	if err := json.Unmarshal([]byte(s), &in); err != nil {
		return Services{}, err
	}

	services := Services{}
	for _, raws := range in {
		for _, raw := range raws {
			service, err := parseService(raw)
			if err != nil {
				return Services{}, err
			}

			services = append(services, service)
		}
	}

	logger.Debug("Services: %s", services)
	return services, nil
}

func parseService(raw json.RawMessage) (Service, error) {
	var service Service

	if err := json.Unmarshal(raw, &service); err != nil {
		return Service{}, err
	}

	return service, nil
}
