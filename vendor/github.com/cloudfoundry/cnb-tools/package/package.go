package package_cnb

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cloudfoundry/libcfbuildpack/helper"
)

const DEFAULT_OS = "linux"

const (
	USAGE         = "Usage:   package.sh <target_os: optional>)\n"
	EXAMPLE       = "Example: package.sh linux\n"
	BUILDPACKTOML = "buildpack.toml"
)

func init() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, USAGE)
		fmt.Fprint(os.Stderr, EXAMPLE)
		os.Exit(0)
	}
}

func Run() error {
	flag.Parse()

	args := len(flag.Args())
	if args > 1 {
		flag.Usage()
	}

	osTarget := flag.Arg(1)
	if osTarget == "" {
		osTarget = DEFAULT_OS
	}

	fmt.Println("Target OS is", osTarget)
	fmt.Println("Creating buildpack directory...")
	currentWorkingPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cwd := filepath.Base(currentWorkingPath)

	timestamp := time.Now().Unix()
	hash := sha256.Sum256([]byte(string(timestamp)))
	guid := hex.EncodeToString(hash[:])

	bpDir := filepath.Join("/tmp", fmt.Sprintf("%s-%s", cwd, guid[:24]))
	if err := os.MkdirAll(bpDir, os.ModePerm); err != nil {
		return err
	}

	fmt.Println("Done")

	fmt.Println("Copying", BUILDPACKTOML+"...")

	if exists, err := helper.FileExists(BUILDPACKTOML); err != nil {
		return err
	} else if exists {
		helper.CopyFile(BUILDPACKTOML, filepath.Join(bpDir, BUILDPACKTOML))
	}
	fmt.Println("Done")

	if err := writeBuildpackTOML(bpDir); err != nil {
		return err
	}

	cmdDir := filepath.Join(currentWorkingPath, "cmd")
	if err := filepath.Walk(cmdDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, "main.go") {
			fmt.Println("Building", path)
			if err := os.MkdirAll(filepath.Join(bpDir, "bin"), os.ModePerm); err != nil {
				return err
			}

			dirName := filepath.Base(filepath.Dir(path))
			cmd := exec.Command("go", "build", "-o", filepath.Join(bpDir, "bin", dirName), path)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("GOOS=%s", osTarget))

			return cmd.Run()
		}
		return nil
	}); err != nil {
		return err
	}

	fmt.Println("Buildpack packaged into:", bpDir)
	return nil
}

func writeBuildpackTOML(bpDir string) error {
	writeFile := filepath.Join(bpDir, BUILDPACKTOML)
	tomlContents, err := ioutil.ReadFile(BUILDPACKTOML)

	if err != nil {
		return err
	}
	bpRewriteHost := os.Getenv("BP_REWRITE_HOST")
	if bpRewriteHost != "" {
		re := regexp.MustCompile(`https:\/\/buildpacks\.cloudfoundry\.org`)
		re.ReplaceAll(tomlContents, []byte("http://"+bpRewriteHost))
	}
	return helper.WriteFile(writeFile, 0777, string(tomlContents))
}
