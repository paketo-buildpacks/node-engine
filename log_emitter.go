package nodeengine

import (
	"io"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type LogEmitter struct {
	// Emitter is embedded and therefore delegates all of its functions to the
	// LogEmitter.
	scribe.Emitter
}

func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Emitter: scribe.NewEmitter(output),
	}
}

func (e LogEmitter) Environment(buildEnv, launchEnv packit.Environment, optimizeMemory bool) {
	e.Process("Configuring build environment")
	e.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(buildEnv))
	e.Break()
	e.Process("Configuring launch environment")
	e.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(launchEnv))
	e.Break()
	e.Subprocess("Writing profile.d/0_memory_available.sh")
	e.Action("Calculates available memory based on container limits at launch time.")
	e.Action("Made available in the MEMORY_AVAILABLE environment variable.")
	if optimizeMemory {
		e.Break()
		e.Subprocess("Writing profile.d/1_optimize_memory.sh")
		e.Action("Assigns the NODE_OPTIONS environment variable with flag setting to optimize memory.")
		e.Action("Limits the total size of all objects on the heap to 75%% of the MEMORY_AVAILABLE.")
	}
	e.Break()
}
