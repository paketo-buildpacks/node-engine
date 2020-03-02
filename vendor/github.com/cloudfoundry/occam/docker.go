package occam

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry/packit/pexec"
)

type Docker struct {
	Image struct {
		Inspect DockerImageInspect
		Remove  DockerImageRemove
	}

	Container struct {
		Inspect DockerContainerInspect
		Logs    DockerContainerLogs
		Run     DockerContainerRun
		Remove  DockerContainerRemove
	}

	Volume struct {
		Remove DockerVolumeRemove
	}
}

func NewDocker() Docker {
	var docker Docker
	executable := pexec.NewExecutable("docker")

	docker.Image.Inspect = DockerImageInspect{executable: executable}
	docker.Image.Remove = DockerImageRemove{executable: executable}

	docker.Container.Inspect = DockerContainerInspect{executable: executable}
	docker.Container.Logs = DockerContainerLogs{executable: executable}
	docker.Container.Run = DockerContainerRun{
		executable: executable,
		inspect:    docker.Container.Inspect,
		env:        map[string]string{"PORT": "8080"},
	}
	docker.Container.Remove = DockerContainerRemove{executable: executable}

	docker.Volume.Remove = DockerVolumeRemove{executable: executable}

	return docker
}

func (d Docker) WithExecutable(executable Executable) Docker {
	d.Image.Inspect.executable = executable
	d.Image.Remove.executable = executable

	d.Container.Inspect.executable = executable
	d.Container.Logs.executable = executable
	d.Container.Remove.executable = executable
	d.Container.Run.executable = executable
	d.Container.Run.inspect = d.Container.Inspect

	d.Volume.Remove.executable = executable

	return d
}

type DockerImageInspect struct {
	executable Executable
}

func (i DockerImageInspect) Execute(ref string) (Image, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := i.executable.Execute(pexec.Execution{
		Args:   []string{"image", "inspect", ref},
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Image{}, fmt.Errorf("failed to inspect docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return NewImageFromInspectOutput(stdout.Bytes())
}

type DockerImageRemove struct {
	executable Executable
}

func (r DockerImageRemove) Execute(ref string) error {
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   []string{"image", "remove", ref},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker image: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerRun struct {
	executable Executable
	inspect    DockerContainerInspect

	command string
	env     map[string]string
	memory  string
	tty     bool
}

func (r DockerContainerRun) WithEnv(env map[string]string) DockerContainerRun {
	r.env = env
	return r
}

func (r DockerContainerRun) WithMemory(memoryLimit string) DockerContainerRun {
	r.memory = memoryLimit
	return r
}

func (r DockerContainerRun) WithCommand(command string) DockerContainerRun {
	r.command = command
	return r
}

func (r DockerContainerRun) WithTTY() DockerContainerRun {
	r.tty = true
	return r
}

func (r DockerContainerRun) Execute(imageID string) (Container, error) {
	args := []string{"container", "run", "--detach"}

	if r.tty {
		args = append(args, "--tty")
	}

	if len(r.env) > 0 {
		var keys []string
		for key := range r.env {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		for _, key := range keys {
			args = append(args, "--env", fmt.Sprintf("%s=%s", key, r.env[key]))
		}
	}

	args = append(args, "--publish", r.env["PORT"], "--publish-all")

	if r.memory != "" {
		args = append(args, "--memory", r.memory)
	}

	args = append(args, imageID)

	if r.command != "" {
		args = append(args, r.command)
	}

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Container{}, fmt.Errorf("failed to run docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return r.inspect.Execute(strings.TrimSpace(stdout.String()))
}

type DockerContainerRemove struct {
	executable Executable
}

func (r DockerContainerRemove) Execute(containerID string) error {
	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   []string{"container", "rm", containerID, "--force"},
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

type DockerContainerInspect struct {
	executable Executable
}

func (i DockerContainerInspect) Execute(containerID string) (Container, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := i.executable.Execute(pexec.Execution{
		Args:   []string{"container", "inspect", containerID},
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return Container{}, fmt.Errorf("failed to inspect docker container: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	container, err := NewContainerFromInspectOutput(stdout.Bytes())
	if err != nil {
		return Container{}, fmt.Errorf("failed to inspect docker container: %w", err)
	}

	return container, nil
}

type DockerContainerLogs struct {
	executable Executable
}

func (i DockerContainerLogs) Execute(containerID string) (fmt.Stringer, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	err := i.executable.Execute(pexec.Execution{
		Args:   []string{"container", "logs", containerID},
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch docker container logs: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout, nil
}

type DockerVolumeRemove struct {
	executable Executable
}

func (r DockerVolumeRemove) Execute(volumes []string) error {
	args := []string{"volume", "rm", "--force"}
	args = append(args, volumes...)

	stderr := bytes.NewBuffer(nil)
	err := r.executable.Execute(pexec.Execution{
		Args:   args,
		Stderr: stderr,
	})
	if err != nil {
		return fmt.Errorf("failed to remove docker volume: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
