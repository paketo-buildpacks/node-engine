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

package test

import (
	"fmt"
	"reflect"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/onsi/gomega/types"
)

// HavePlans tests that a set of plans is returned from detect.
func HavePlans(plans ...buildplan.Plan) types.GomegaMatcher {
	return &havePlans{plans}
}

type havePlans struct {
	plans []buildplan.Plan
}

func (p *havePlans) Match(actual interface{}) (bool, error) {
	a, ok := actual.(buildplan.Plans)
	if !ok {
		return false, fmt.Errorf("actual must be a buildplan.Plans: %s", reflect.TypeOf(a))
	}

	return reflect.DeepEqual(a, p.expected()), nil
}

func (p *havePlans) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\n to match\n\t%#v\n", actual, p.expected())
}

func (p *havePlans) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\n not to match\n\t%#v\n", actual, p.expected())
}

func (p *havePlans) expected() buildplan.Plans {
	e := buildplan.Plans{}

	if len(p.plans) > 0 {
		e.Plan = p.plans[0]
	}

	if len(p.plans) > 1 {
		e.Or = p.plans[1:]
	}

	return e
}
