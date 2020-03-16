package node

import (
	"sort"

	"github.com/cloudfoundry/packit"
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

	sort.Slice(entries, func(i, j int) bool {
		leftSource := entries[i].Metadata["version-source"]
		left, _ := leftSource.(string)

		rightSource := entries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
	})

	chosenEntry := entries[0]

	if chosenEntry.Metadata == nil {
		chosenEntry.Metadata = map[string]interface{}{}
	}

	for _, entry := range entries {
		if entry.Metadata["build"] == true {
			chosenEntry.Metadata["build"] = true
		}
	}

	r.logger.Candidates(entries)

	return chosenEntry
}
