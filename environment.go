package nodeengine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

var (
	MemoryAvailableScript = strings.TrimSpace(`if [ -z "$MEMORY_AVAILABLE" ]; then
  if [ -f "/sys/fs/cgroup/cgroup.controllers" ]; then
    memory_in_bytes="$(cat /sys/fs/cgroup/memory.max)"
  else
    memory_in_bytes="$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)"
  fi
  if [ "$memory_in_bytes" != "" ] && [ "$memory_in_bytes" != "max" ]; then
    MEMORY_AVAILABLE="$((memory_in_bytes / (1024 * 1024)))"
    export MEMORY_AVAILABLE
  fi
fi
`)

	OptimizeMemoryScript = `if [ -n "$MEMORY_AVAILABLE" ]; then
  export NODE_OPTIONS="--max_old_space_size=$(( MEMORY_AVAILABLE * 75 / 100 ))"
fi
`)

type Environment struct {
	logger LogEmitter
}

func NewEnvironment(logger LogEmitter) Environment {
	return Environment{
		logger: logger,
	}
}

func (e Environment) Configure(buildEnv, launchEnv packit.Environment, path string, optimizeMemory bool) error {
	launchEnv.Default("NODE_HOME", path)
	launchEnv.Default("NODE_ENV", "production")
	launchEnv.Default("NODE_VERBOSE", "false")

	buildEnv.Default("NODE_HOME", path)
	if val, ok := os.LookupEnv("NODE_ENV"); ok {
		buildEnv.Default("NODE_ENV", val)
	} else {
		buildEnv.Default("NODE_ENV", "production")
	}

	if val, ok := os.LookupEnv("NODE_VERBOSE"); ok {
		buildEnv.Default("NODE_VERBOSE", val)
	} else {
		buildEnv.Default("NODE_VERBOSE", "false")
	}

	profileDPath := filepath.Join(path, "profile.d")
	err := os.MkdirAll(profileDPath, os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(profileDPath, "0_memory_available.sh"), []byte(MemoryAvailableScript), 0644)
	if err != nil {
		return err
	}

	if optimizeMemory {
		err = ioutil.WriteFile(filepath.Join(profileDPath, "1_optimize_memory.sh"), []byte(OptimizeMemoryScript), 0644)
		if err != nil {
			return err
		}
	}

	e.logger.Environment(buildEnv, launchEnv, optimizeMemory)

	return nil
}
