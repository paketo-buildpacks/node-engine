package nodeengine

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
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

type Environment struct {
	logger LogEmitter
}

func NewEnvironment(logger LogEmitter) Environment {
	return Environment{
		logger: logger,
	}
}

func (e Environment) Configure(env packit.Environment, path string, optimizeMemory bool) error {
	env.Default("NODE_HOME", path)
	if val, ok := os.LookupEnv("NODE_ENV"); ok {
		env.Default("NODE_ENV", val)
	} else {
		env.Default("NODE_ENV", "production")
	}

	if val, ok := os.LookupEnv("NODE_VERBOSE"); ok {
		env.Default("NODE_VERBOSE", val)
	} else {
		env.Default("NODE_VERBOSE", "false")
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

	e.logger.Environment(env, optimizeMemory)

	return nil
}
