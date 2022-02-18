package nodeengine

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
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
