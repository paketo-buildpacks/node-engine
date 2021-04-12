package nodeengine

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type NodeVersionParser struct{}

func NewNodeVersionParser() NodeVersionParser {
	return NodeVersionParser{}
}

func (p NodeVersionParser) ParseVersion(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	version, err := p.validateNodeVersion(string(content))
	if err != nil {
		return "", err
	}

	return version, nil
}

func (p NodeVersionParser) validateNodeVersion(content string) (string, error) {
	content = strings.TrimSpace(strings.ToLower(content))

	content = strings.TrimPrefix(content, "v")

	if _, err := semver.NewConstraint(content); err != nil {
		return "", fmt.Errorf("invalid version constraint specified in .node-version: %q", content)
	}

	return content, nil
}
