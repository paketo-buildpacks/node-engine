package internal

import "strings"

func LoadEnvironmentMap(environ []string) map[string]string {
	environment := map[string]string{}

	for _, variable := range environ {
		parts := strings.SplitN(variable, "=", 2)
		environment[parts[0]] = parts[1]
	}

	return environment
}
