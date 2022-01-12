package nodeengine

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	lts = map[string]int{
		"argon":   4,
		"boron":   6,
		"carbon":  8,
		"dubnium": 10,
		"erbium":  12,
		"fermium": 14,
		"gallium": 16,
	}
)

type NvmrcParser struct{}

func NewNvmrcParser() NvmrcParser {
	return NvmrcParser{}
}

func (p NvmrcParser) ParseVersion(path string) (string, error) {
	nvmrcContents, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	nvmrcVersion, err := p.validateNvmrc(string(nvmrcContents))
	if err != nil {
		return "", err
	}

	// TODO
	// if nvmrcVersion == Node {
	// 	logger.Info(".nvmrc specified latest node version, this will be selected from versions available in buildpack.toml")
	// }
	//
	// if strings.HasPrefix(nvmrcVersion, "lts") {
	// 	logger.Info(".nvmrc specified an lts version, this will be selected from versions available in buildpack.toml")
	// }

	return p.formatNvmrcContent(nvmrcVersion), nil
}

func (p NvmrcParser) validateNvmrc(content string) (string, error) {
	content = strings.TrimSpace(strings.ToLower(content))

	if content == "lts/*" || content == Node {
		return content, nil
	}

	for key := range lts {
		if content == strings.ToLower("lts/"+key) {
			return content, nil
		}
	}

	content = strings.TrimPrefix(content, "v")

	if _, err := semver.NewConstraint(content); err != nil {
		return "", fmt.Errorf("invalid version constraint specified in .nvmrc: %q", content)
	}

	return content, nil
}

func (p NvmrcParser) formatNvmrcContent(version string) string {
	if version == Node {
		return "*"
	}

	if strings.HasPrefix(version, "lts") {
		ltsName := strings.SplitN(version, "/", 2)[1]
		if ltsName == "*" {
			var maxVersion int
			for _, versionValue := range lts {
				if maxVersion < versionValue {
					maxVersion = versionValue
				}
			}

			return fmt.Sprintf("%d.*", maxVersion)
		}

		return fmt.Sprintf("%d.*", lts[ltsName])
	}

	return version
}
