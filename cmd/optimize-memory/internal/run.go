package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

func Run(environment map[string]string, output io.Writer, root string) error {
	if _, ok := environment["MEMORY_AVAILABLE"]; !ok {
		var path string

		_, err := os.Stat(filepath.Join(root, "sys", "fs", "cgroup", "cgroup.controllers"))
		switch {
		case err == nil:
			path = filepath.Join(root, "sys", "fs", "cgroup", "memory.max")
		case errors.Is(err, os.ErrNotExist):
			path = filepath.Join(root, "sys", "fs", "cgroup", "memory", "memory.limit_in_bytes")
		default:
			return err
		}

		if path != "" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			environment["MEMORY_AVAILABLE"] = strings.TrimSpace(string(content))
		}
	}

	variables := map[string]string{}
	if environment["MEMORY_AVAILABLE"] != "" && environment["MEMORY_AVAILABLE"] != "max" {
		memory, err := strconv.Atoi(environment["MEMORY_AVAILABLE"])
		if err != nil {
			return err
		}

		variables["MEMORY_AVAILABLE"] = strconv.Itoa(memory / (1024 * 1024))
	}

	if _, ok := environment["OPTIMIZE_MEMORY"]; ok {
		if _, ok := variables["MEMORY_AVAILABLE"]; ok {
			memoryMax, err := strconv.Atoi(variables["MEMORY_AVAILABLE"])
			if err != nil {
				return err
			}

			variables["NODE_OPTIONS"] = fmt.Sprintf("--max_old_space_size=%d", memoryMax*75/100)
		}
	}

	return toml.NewEncoder(output).Encode(variables)
}
