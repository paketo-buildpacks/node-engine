package matchers

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/occam"
	"github.com/onsi/gomega/types"
)

func BeAvailable() types.GomegaMatcher {
	return &BeAvailableMatcher{}
}

type BeAvailableMatcher struct {
}

func (*BeAvailableMatcher) Match(actual interface{}) (bool, error) {
	container, ok := actual.(occam.Container)
	if !ok {
		return false, fmt.Errorf("BeAvailableMatcher expects an occam.Container, received %T", actual)
	}

	response, err := http.Get(fmt.Sprintf("http://localhost:%s", container.HostPort()))
	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	return true, nil
}

func (*BeAvailableMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nto be available", actual)
}

func (*BeAvailableMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be available", actual)
}
