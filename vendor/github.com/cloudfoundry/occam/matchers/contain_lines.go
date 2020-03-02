package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func ContainLines(expected ...interface{}) types.GomegaMatcher {
	return &containLinesMatcher{
		expected: expected,
	}
}

type containLinesMatcher struct {
	expected interface{}
}

func (matcher *containLinesMatcher) Match(actual interface{}) (success bool, err error) {
	raw, ok := actual.(string)
	if !ok {
		stringer, ok := actual.(fmt.Stringer)
		if !ok {
			return false, fmt.Errorf("actual is not a string or fmt.Stringer: %#v", actual)
		}
		raw = stringer.String()
	}

	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "[builder]") {
			lines = append(lines, strings.TrimPrefix(line, "[builder] "))
		}
	}

	expectedLength := reflect.ValueOf(matcher.expected).Len()
	actualLength := len(lines)
	for i := 0; i < (actualLength - expectedLength + 1); i++ {
		eSlice := reflect.ValueOf(matcher.expected).Slice(0, expectedLength)

		match := true
		for j := 0; j < eSlice.Len(); j++ {
			aValue := lines[j]
			eValue := eSlice.Index(j)

			if eMatcher, ok := eValue.Interface().(types.GomegaMatcher); ok {
				m, err := eMatcher.Match(aValue)
				if err != nil {
					return false, err
				}

				if !m {
					match = false
				}
			} else if !reflect.DeepEqual(aValue, eValue.Interface()) {
				match = false
			}
		}

		if match {
			return true, nil
		}
	}

	return false, nil
}

func (matcher *containLinesMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(matcher.lines(actual), "to contain lines", matcher.expected)
}

func (matcher *containLinesMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(matcher.lines(actual), "not to contain lines", matcher.expected)
}

func (matcher *containLinesMatcher) lines(actual interface{}) []string {
	raw, ok := actual.(string)
	if !ok {
		raw = actual.(fmt.Stringer).String()
	}

	var lines []string
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "[builder]") {
			lines = append(lines, strings.TrimPrefix(line, "[builder] "))
		}
	}

	return lines
}
