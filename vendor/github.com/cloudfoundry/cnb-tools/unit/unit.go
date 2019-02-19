package unit_cnb

import (
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/cnb-tools/utils"
)

func Run() error {
	fmt.Println("Run Buildpack Unit Tests")

	cmd := exec.Command("go", "test", "./...", "-v", "-run", "Unit")
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))

	if err != nil {
		fmt.Printf(utils.RED, "GO Test Failed")
		return err
	} else {
		fmt.Printf(utils.GREEN, "GO Test Succeeded")
	}
	return nil
}
