package node

import (
	"io"
	"strconv"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

type LogEmitter struct {
	// Logger is embedded and therefore delegates all of its functions to the
	// LogEmitter.
	scribe.Logger
}

func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Logger: scribe.NewLogger(output),
	}
}

func (e LogEmitter) SelectedDependency(entry packit.BuildpackPlanEntry, dependency postal.Dependency, now time.Time) {
	source, ok := entry.Metadata["version-source"].(string)
	if !ok {
		source = "<unknown>"
	}

	e.Subprocess("Selected %s version (using %s): %s", dependency.Name, source, dependency.Version)

	if (dependency.DeprecationDate != time.Time{}) {
		deprecationDate := dependency.DeprecationDate
		switch {
		case (deprecationDate.Add(-30*24*time.Hour).Before(now) && deprecationDate.After(now)):
			e.Action("Version %s of %s will be deprecated after %s.", dependency.Version, dependency.Name, dependency.DeprecationDate.Format("2006-01-02"))
			e.Action("Migrate your application to a supported version of %s before this time.", dependency.Name)
		case (deprecationDate == now || deprecationDate.Before(now)):
			e.Action("Version %s of %s is deprecated.", dependency.Version, dependency.Name)
			e.Action("Migrate your application to a supported version of %s.", dependency.Name)
		}
	}
	e.Break()
}

func (e LogEmitter) Environment(env packit.Environment, optimizeMemory bool) {
	e.Process("Configuring environment")
	e.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(env))
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

func (e LogEmitter) Candidates(entries []packit.BuildpackPlanEntry) {
	e.Subprocess("Candidate version sources (in priority order):")

	var (
		sources [][2]string
		maxLen  int
	)

	for _, entry := range entries {
		versionSource, ok := entry.Metadata["version-source"].(string)
		if !ok {
			versionSource = "<unknown>"
		}

		if len(versionSource) > maxLen {
			maxLen = len(versionSource)
		}

		if entry.Version == "" {
			entry.Version = "*"
		}

		sources = append(sources, [2]string{versionSource, entry.Version})
	}

	for _, source := range sources {
		e.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	e.Break()
}
