package node

import (
	"errors"

	"github.com/cloudfoundry/packit"
)

type CacheHandler struct{}

func NewCacheHandler() CacheHandler {
	return CacheHandler{}
}

func (ch CacheHandler) Match(layer packit.Layer, dependency BuildpackMetadataDependency) (bool, error) {
	contents, exists := layer.Metadata[DepKey]
	if !exists {
		return false, nil
	}
	contentString, ok := contents.(string)
	if !ok {
		return false, errors.New("layer metadata dependency-sha was not a string")
	}
	return contentString == dependency.SHA256, nil
}
