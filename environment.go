package nodeengine

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

var (
	MemoryAvailableScript = strings.TrimSpace(`set -e
if [ -z "$MEMORY_AVAILABLE" ]; then
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

	OptimizeMemoryScript = `set -e
if [ -n "$MEMORY_AVAILABLE" ]; then
  export NODE_OPTIONS="--max_old_space_size=$(( MEMORY_AVAILABLE * 75 / 100 ))"
fi
`
)

type Environment struct {
	logger LogEmitter
}

func NewEnvironment(logger LogEmitter) Environment {
	return Environment{
		logger: logger,
	}
}

func (e Environment) Configure(buildEnv, launchEnv packit.Environment, layerPath, execdPath string, optimizeMemory bool) error {
	launchEnv.Default("NODE_HOME", layerPath)
	launchEnv.Default("NODE_ENV", "production")
	launchEnv.Default("NODE_VERBOSE", "false")

	buildEnv.Default("NODE_HOME", layerPath)
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

	execdDir := filepath.Join(layerPath, "exec.d")
	err := os.MkdirAll(execdDir, os.ModePerm)
	if err != nil {
		return err
	}

	err = fs.Copy(execdPath, filepath.Join(execdDir, "0-optimize-memory"))
	if err != nil {
		return err
	}

	if optimizeMemory {
		launchEnv.Default("OPTIMIZE_MEMORY", "true")
	}

	e.logger.Environment(buildEnv, launchEnv, optimizeMemory)

	return nil
}
