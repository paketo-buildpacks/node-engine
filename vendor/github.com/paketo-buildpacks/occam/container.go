package occam

import (
	"encoding/json"
	"strings"
)

type Container struct {
	ID    string
	Ports map[string]string
	Env   map[string]string
}

func NewContainerFromInspectOutput(output []byte) (Container, error) {
	var inspect []struct {
		ID     string `json:"Id"`
		Config struct {
			Env []string `json:"Env"`
		} `json:"Config"`
		NetworkSettings struct {
			Ports map[string][]struct {
				HostPort string `json:"HostPort"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}

	err := json.Unmarshal(output, &inspect)
	if err != nil {
		return Container{}, err
	}

	container := Container{ID: inspect[0].ID}

	if len(inspect[0].NetworkSettings.Ports) > 0 {
		container.Ports = make(map[string]string)

		for key, value := range inspect[0].NetworkSettings.Ports {
			container.Ports[strings.TrimSuffix(key, "/tcp")] = value[0].HostPort
		}
	}

	if len(inspect[0].Config.Env) > 0 {
		container.Env = make(map[string]string)

		for _, e := range inspect[0].Config.Env {
			parts := strings.SplitN(e, "=", 2)
			container.Env[parts[0]] = parts[1]
		}
	}

	return container, nil
}

func (c Container) HostPort() string {
	return c.Ports[c.Env["PORT"]]
}
