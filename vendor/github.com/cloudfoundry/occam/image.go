package occam

import (
	"encoding/json"
	"fmt"
)

type Image struct {
	ID         string
	Buildpacks []ImageBuildpackMetadata
}

type ImageBuildpackMetadata struct {
	Key    string
	Layers map[string]ImageBuildpackMetadataLayer
}

type ImageBuildpackMetadataLayer struct {
	SHA      string
	Build    bool
	Launch   bool
	Cache    bool
	Metadata map[string]interface{}
}

func NewImageFromInspectOutput(output []byte) (Image, error) {
	var inspect []struct {
		ID     string `json:"Id"`
		Config struct {
			Labels struct {
				LifecycleMetadata string `json:"io.buildpacks.lifecycle.metadata"`
			} `json:"Labels"`
		} `json:"Config"`
	}
	err := json.Unmarshal(output, &inspect)
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w", err)
	}

	var metadata struct {
		Buildpacks []struct {
			Key    string `json:"key"`
			Layers map[string]struct {
				SHA    string                 `json:"sha"`
				Build  bool                   `json:"build"`
				Launch bool                   `json:"launch"`
				Cache  bool                   `json:"cache"`
				Data   map[string]interface{} `json:"data"`
			} `json:"layers"`
		} `json:"buildpacks"`
	}
	err = json.Unmarshal([]byte(inspect[0].Config.Labels.LifecycleMetadata), &metadata)
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w", err)
	}

	var buildpacks []ImageBuildpackMetadata
	for _, buildpack := range metadata.Buildpacks {
		layers := map[string]ImageBuildpackMetadataLayer{}
		for name, layer := range buildpack.Layers {
			layers[name] = ImageBuildpackMetadataLayer{
				SHA:      layer.SHA,
				Build:    layer.Build,
				Launch:   layer.Launch,
				Cache:    layer.Cache,
				Metadata: layer.Data,
			}
		}

		buildpacks = append(buildpacks, ImageBuildpackMetadata{
			Key:    buildpack.Key,
			Layers: layers})
	}

	return Image{
		ID:         inspect[0].ID,
		Buildpacks: buildpacks,
	}, nil
}
