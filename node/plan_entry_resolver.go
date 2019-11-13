package node

import (
	"sort"
	"strconv"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/scribe"
)

type PlanEntryResolver struct {
	logger scribe.Logger
}

func NewPlanEntryResolver(logger scribe.Logger) PlanEntryResolver {
	return PlanEntryResolver{
		logger: logger,
	}
}

func (r PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	var (
		priorities = map[string]int{
			"buildpack.yml": 3,
			"package.json":  2,
			".nvmrc":        1,
			"":              -1,
		}
	)

	sort.Slice(entries, func(i, j int) bool {
		leftSource := entries[i].Metadata["version-source"]
		left, _ := leftSource.(string)

		rightSource := entries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
	})

	chosenEntry := entries[0]

	r.printCandidates(entries)

	if chosenEntry.Metadata == nil {
		chosenEntry.Metadata = map[string]interface{}{}
	}

	for _, entry := range entries {
		if entry.Metadata["build"] == true {
			chosenEntry.Metadata["build"] = true
		}

		if entry.Metadata["cache"] == true {
			chosenEntry.Metadata["cache"] = true
		}
	}

	return chosenEntry
}

func (r PlanEntryResolver) printCandidates(entries []packit.BuildpackPlanEntry) {
	r.logger.Subprocess("Candidate version sources (in priority order):")

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

		sources = append(sources, [2]string{versionSource, entry.Version})
	}

	for _, source := range sources {
		r.logger.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	r.logger.Break()
}
