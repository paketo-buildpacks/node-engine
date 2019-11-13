package node

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/packit/scribe"
)

var (
	MemoryAvailableScript = strings.TrimSpace(`
if [[ -z "$MEMORY_AVAILABLE" ]]; then
		memory_in_bytes="$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)"
		MEMORY_AVAILABLE="$(( $memory_in_bytes / ( 1024 * 1024 ) ))"
fi
export MEMORY_AVAILABLE
`)

	OptimizeMemoryScript = `export NODE_OPTIONS="--max_old_space_size=$(( $MEMORY_AVAILABLE * 75 / 100 ))"`
)

//go:generate faux --interface EnvironmentVariables --output fakes/environment_variables.go
type EnvironmentVariables interface {
	Override(key, value string)
}

type Environment struct {
	logger scribe.Logger
}

func NewEnvironment(logger scribe.Logger) Environment {
	return Environment{
		logger: logger,
	}
}

func (e Environment) Configure(env EnvironmentVariables, path string, optimizeMemory bool) error {
	e.logger.Subprocess("Configuring environment")

	envMap := scribe.FormattedMap{
		"NODE_HOME":    path,
		"NODE_ENV":     "production",
		"NODE_VERBOSE": "false",
	}

	for key, val := range envMap {
		env.Override(key, val.(string))
	}

	e.logger.Action("%s", envMap)
	e.logger.Break()

	profileDPath := filepath.Join(path, "profile.d")
	err := os.MkdirAll(profileDPath, os.ModePerm)
	if err != nil {
		return err
	}

	e.logger.Action("Writing profile.d/0_memory_available.sh")
	e.logger.Detail("Calculates available memory based on container limits at launch time.")
	e.logger.Detail("Made available in the MEMORY_AVAILABLE environment variable.")

	err = ioutil.WriteFile(filepath.Join(profileDPath, "0_memory_available.sh"), []byte(MemoryAvailableScript), 0644)
	if err != nil {
		return err
	}

	if optimizeMemory {
		e.logger.Break()
		e.logger.Action("Writing profile.d/1_optimize_memory.sh")
		e.logger.Detail("Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.")
		e.logger.Detail("Limits the total size of all objects on the heap to 75%% of the MEMORY_AVAILABLE.")

		err = ioutil.WriteFile(filepath.Join(profileDPath, "1_optimize_memory.sh"), []byte(OptimizeMemoryScript), 0644)
		if err != nil {
			return err
		}
	}

	e.logger.Break()

	return nil
}
