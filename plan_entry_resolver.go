package nodeengine

import (
	"sort"

	"github.com/paketo-buildpacks/packit"
)

type PlanEntryResolver struct {
	logger LogEmitter
}

func NewPlanEntryResolver(logger LogEmitter) PlanEntryResolver {
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

	filteredEntries := []packit.BuildpackPlanEntry{}
	// The metadata from "npm" entry are not relevant
	for _, e := range entries {
		if e.Name == Node {
			filteredEntries = append(filteredEntries, e)
		}
	}

	sort.Slice(filteredEntries, func(i, j int) bool {
		leftSource := filteredEntries[i].Metadata["version-source"]
		left, _ := leftSource.(string)

		rightSource := filteredEntries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
	})

	chosenEntry := filteredEntries[0]

	if chosenEntry.Metadata == nil {
		chosenEntry.Metadata = map[string]interface{}{}
	}

	for _, entry := range filteredEntries {
		if entry.Metadata["build"] == true {
			chosenEntry.Metadata["build"] = true
		}

		if entry.Metadata["launch"] == true {
			chosenEntry.Metadata["launch"] = true
		}
	}

	r.logger.Candidates(filteredEntries)

	return chosenEntry
}
