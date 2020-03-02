package occam

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cloudfoundry/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

//go:generate faux --interface DockerImageInspectClient --output fakes/docker_image_inspect_client.go
type DockerImageInspectClient interface {
	Execute(ref string) (Image, error)
}

type Pack struct {
	Build PackBuild
}

func NewPack() Pack {
	return Pack{
		Build: PackBuild{
			executable:               pexec.NewExecutable("pack"),
			dockerImageInspectClient: NewDocker().Image.Inspect,
		},
	}
}

func (p Pack) WithExecutable(executable Executable) Pack {
	p.Build.executable = executable
	return p
}

func (p Pack) WithDockerImageInspectClient(client DockerImageInspectClient) Pack {
	p.Build.dockerImageInspectClient = client
	return p
}

func (p Pack) WithVerbose() Pack {
	p.Build.verbose = true
	return p
}

func (p Pack) WithNoColor() Pack {
	p.Build.noColor = true
	return p
}

type PackBuild struct {
	executable               Executable
	dockerImageInspectClient DockerImageInspectClient

	verbose bool
	noColor bool

	buildpacks []string
	network    string
	builder    string
	clearCache bool
	env        map[string]string
	noPull     bool
}

func (pb PackBuild) WithBuildpacks(buildpacks ...string) PackBuild {
	pb.buildpacks = append(pb.buildpacks, buildpacks...)
	return pb
}

func (pb PackBuild) WithNetwork(name string) PackBuild {
	pb.network = name
	return pb
}

func (pb PackBuild) WithBuilder(name string) PackBuild {
	pb.builder = name
	return pb
}

func (pb PackBuild) WithClearCache() PackBuild {
	pb.clearCache = true
	return pb
}

func (pb PackBuild) WithEnv(env map[string]string) PackBuild {
	pb.env = env
	return pb
}

func (pb PackBuild) WithNoPull() PackBuild {
	pb.noPull = true
	return pb
}

func (pb PackBuild) Execute(name, path string) (Image, fmt.Stringer, error) {
	args := []string{"build", name}

	if pb.verbose {
		args = append(args, "--verbose")
	}

	if pb.noColor {
		args = append(args, "--no-color")
	}

	args = append(args, "--path", path)

	for _, buildpack := range pb.buildpacks {
		args = append(args, "--buildpack", buildpack)
	}

	if pb.network != "" {
		args = append(args, "--network", pb.network)
	}

	if pb.builder != "" {
		args = append(args, "--builder", pb.builder)
	}

	if pb.clearCache {
		args = append(args, "--clear-cache")
	}

	if len(pb.env) != 0 {
		var variables []string
		for key, value := range pb.env {
			variables = append(variables, fmt.Sprintf("%s=%s", key, value))
		}

		sort.Strings(variables)

		for _, v := range variables {
			args = append(args, "--env", v)
		}
	}

	if pb.noPull {
		args = append(args, "--no-pull")
	}

	buildLogBuffer := bytes.NewBuffer(nil)
	err := pb.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: buildLogBuffer,
		Stderr: buildLogBuffer,
	})
	if err != nil {
		return Image{}, buildLogBuffer, fmt.Errorf("failed to pack build: %w\n\nOutput:\n%s", err, buildLogBuffer)
	}

	image, err := pb.dockerImageInspectClient.Execute(name)
	if err != nil {
		return Image{}, buildLogBuffer, fmt.Errorf("failed to pack build: %w", err)
	}

	return image, buildLogBuffer, nil
}
