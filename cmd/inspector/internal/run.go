package internal

import (
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

func Run(environment map[string]string, output io.Writer, root string) error {
	variables := map[string]string{}
	if debug, ok := environment["BPL_DEBUG_ENABLED"]; ok && debug == "true" {
		option := "--inspect"
		if debugPort, ok := environment["BPL_DEBUG_PORT"]; ok {
			option = fmt.Sprintf("%s=127.0.0.1:%s", option, debugPort)
		}

		if nodeOpts, ok := environment["NODE_OPTIONS"]; ok {
			if strings.Contains(nodeOpts, "--inspect") {
				return nil
			}

			option = strings.Join([]string{nodeOpts, option}, " ")
		}

		variables["NODE_OPTIONS"] = option
	}

	return toml.NewEncoder(output).Encode(variables)
}
