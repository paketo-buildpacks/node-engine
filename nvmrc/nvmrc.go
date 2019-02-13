package nvmrc

import (
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

var lts = map[string]int{
	"argon":   4,
	"boron":   6,
	"carbon":  8,
	"dubnium": 10,
}

type Logger interface {
	Info(format string, args ...interface{})
}

func GetVersion(path string, logger Logger) (string, error) {
	nvmrcContents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	nvmrcVersion, err := validateNvmrc(string(nvmrcContents))
	if err != nil {
		return "", err
	}

	if nvmrcVersion == "node" {
		logger.Info(".nvmrc specified latest node version, this will be selected from versions available in buildpack.toml")
	}

	if strings.HasPrefix(nvmrcVersion, "lts") {
		logger.Info(".nvmrc specified an lts version, this will be selected from versions available in buildpack.toml")
	}

	return formatNvmrcContent(nvmrcVersion), nil
}

func validateNvmrc(content string) (string, error) {
	content = strings.TrimSpace(strings.ToLower(content))

	if content == "lts/*" || content == "node" {
		return content, nil
	}

	for key, _ := range lts {
		if content == strings.ToLower("lts/"+key) {
			return content, nil
		}
	}

	if len(content) > 0 && content[0] == 'v' {
		content = content[1:]
	}

	if _, err := semver.NewVersion(content); err != nil {
		return "", fmt.Errorf("invalid version %s specified in .nvmrc", err)
	}

	return content, nil
}

func formatNvmrcContent(version string) string {
	if version == "node" {
		return "*"
	} else if strings.HasPrefix(version, "lts") {
		ltsName := strings.Split(version, "/")[1]
		if ltsName == "*" {
			maxVersion := 0
			for _, versionValue := range lts {
				if maxVersion < versionValue {
					maxVersion = versionValue
				}
			}
			return strconv.Itoa(maxVersion) + ".*.*"
		} else {
			versionNumber := lts[ltsName]
			return strconv.Itoa(versionNumber) + ".*.*"
		}
	} else {
		matcher := regexp.MustCompile(semver.SemVerRegex)

		groups := matcher.FindStringSubmatch(version)
		for index := 0; index < len(groups); index++ {
			if groups[index] == "" {
				groups = append(groups[:index], groups[index+1:]...)
				index--
			}
		}

		return version + strings.Repeat(".*", 4-len(groups))
	}
}
