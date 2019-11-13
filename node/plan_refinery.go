package node

import "github.com/cloudfoundry/packit"

type PlanRefiner struct{}

func NewPlanRefiner() PlanRefiner {
	return PlanRefiner{}
}

func (pf PlanRefiner) BillOfMaterial(dependency BuildpackMetadataDependency) packit.BuildpackPlan {
	return packit.BuildpackPlan{
		Entries: []packit.BuildpackPlanEntry{
			{
				Name:    dependency.ID,
				Version: dependency.Version,
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     dependency.Name,
					"sha256":   dependency.SHA256,
					"stacks":   dependency.Stacks,
					"uri":      dependency.URI,
				},
			},
		},
	}
}
