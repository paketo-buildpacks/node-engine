package nodeengine

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/semver"
)

type NodeVersionParser struct{}

func NewNodeVersionParser() NodeVersionParser {
	return NodeVersionParser{}
}

func (p NodeVersionParser) ParseVersion(path string) (string, error) {
	nodeVersionContents, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	nodeVersion, err := p.validateNodeVersion(string(nodeVersionContents))
	if err != nil {
		return "", err
	}

	return p.formatNodeVersionContent(nodeVersion), nil
}

func (p NodeVersionParser) validateNodeVersion(content string) (string, error) {
	content = strings.TrimSpace(strings.ToLower(content))

	content = strings.TrimPrefix(content, "v")

	if _, err := semver.NewVersion(content); err != nil {
		return "", fmt.Errorf("invalid version specified in .node-version: %q", content)
	}

	return content, nil
}

func (p NodeVersionParser) formatNodeVersionContent(version string) string {
	var groups []string
	for _, part := range semverRegex.FindStringSubmatch(version) {
		if part != "" {
			groups = append(groups, part)
		}
	}

	return version + strings.Repeat(".*", 4-len(groups))
}
