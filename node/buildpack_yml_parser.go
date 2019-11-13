package node

import (
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Version         string `yaml:"version"`
	OptimizedMemory bool   `yaml:"optimize-memory"`
}

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) Parse(path string) (Config, error) {
	var buildpack struct {
		NodeJS Config `yaml:"nodejs"`
	}

	file, err := os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}
	defer file.Close()

	if !os.IsNotExist(err) {
		err = yaml.NewDecoder(file).Decode(&buildpack)
		if err != nil {
			return Config{}, err
		}
	}

	if os.Getenv("OPTIMIZE_MEMORY") == "true" {
		buildpack.NodeJS.OptimizedMemory = true
	}

	return buildpack.NodeJS, nil
}

func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	config, err := p.Parse(path)
	if err != nil {
		return "", err
	}

	return config.Version, nil
}
