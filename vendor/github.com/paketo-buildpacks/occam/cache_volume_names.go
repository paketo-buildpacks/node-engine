package occam

import (
	"crypto/sha256"
	"fmt"
)

func CacheVolumeNames(name string) []string {
	refName := []byte(fmt.Sprintf("index.docker.io/library/%s:latest", name))

	sum := sha256.Sum256(refName)

	var volumes []string
	for _, t := range []string{"build", "launch", "cache"} {
		volumes = append(volumes, fmt.Sprintf("pack-cache-%x.%s", sum[:6], t))
	}

	return volumes
}
