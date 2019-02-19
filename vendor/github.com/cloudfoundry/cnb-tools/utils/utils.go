package utils

import (
	"os/exec"
	"syscall"
)

const (
	RED   = "\n\033[0;31m%s\033[0m\n"
	GREEN = "\n\033[0;32m%s\033[0m\n"
)

func ExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}
