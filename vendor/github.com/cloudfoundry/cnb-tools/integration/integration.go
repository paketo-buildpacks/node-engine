package integration_cnb

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudfoundry/cnb-tools/utils"

	"github.com/cloudfoundry/cnb-tools/action"
)

const (
	INTEGRATION         = "integration"
	ENVPACK             = "PACK_VERSION"
	DEFAULT_BUILD_IMAGE = "cfbuildpacks/cflinuxfs3-cnb-experimental:build"
	DEFAULT_RUN_IMAGE   = "cfbuildpacks/cflinuxfs3-cnb-experimental:run"
)

func Run() error {
	if _, err := os.Stat(INTEGRATION); os.IsNotExist(err) {
		fmt.Println("** WARNING ** No integration tests specified")
		return nil
	}

	envPack := os.Getenv(ENVPACK)
	if envPack != "" {
		fmt.Println("Using the", ENVPACK, "environment variable")
	}
	action.InstallTools(envPack)

	buildImage := os.Getenv("CNB_BUILD_IMAGE")
	if buildImage == "" {
		buildImage = DEFAULT_BUILD_IMAGE
	}

	runImage := os.Getenv("CNB_RUN_IMAGE")
	if runImage == "" {
		runImage = DEFAULT_RUN_IMAGE
	}

	for _, image := range []string{runImage, buildImage} {
		cmd := exec.Command("docker", "pull", image)
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	fmt.Println("Run Buildpack Runtime Integration Tests")

	cmd := exec.Command("go", "test", "-timeout", "0", "./integration/...", "-v", "-run", "Integration")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	if err != nil {
		fmt.Printf(utils.RED, "GO Test Failed")
		return err
	} else {
		fmt.Printf(utils.GREEN, "GO Test Succeeded")
	}
	return nil
}
