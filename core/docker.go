package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/dustinliu/devspace/logging"
)

var docker_exec string

func init() {
	var err error
	docker_exec, err = exec.LookPath("docker")
	if err != nil {
		logging.Fatal(err)
	}
}

type RunOptions struct {
	Mount   map[string]string
	Command []string
	Rm      bool
	Env     map[string]string
	Fork    bool
	Detach  bool
}

type ExecOptions struct {
	Fork    bool
	Tty     bool
	workDir string
}

type Docker interface {
	ListImages() ([]types.ImageSummary, error)
	BuildImage(tag, dockerfile, path string) error
	ListContains() ([]types.Container, error)
	Run(imageID, containerName string, opt RunOptions) error
	Attach(container types.Container) error
	Exec(containerName string, cmd []string, opt ExecOptions) error
}

func newDocker() *DockerAPI {
	return &DockerAPI{}
}

type DockerAPI struct{}

func (d *DockerAPI) BuildImage(tag, dockerfile, path string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer tar.Close()

	opts := types.ImageBuildOptions{
		Dockerfile: dockerfile,
		Tags:       []string{tag},
		Remove:     true,
		PullParent: false,
	}
	response, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	type buildResponse struct {
		Stream string `json:"stream"`
	}
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		r := buildResponse{}
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			return err
		}
		if r.Stream != "" {
			fmt.Print(r.Stream)
		}
	}

	return nil
}

func (d *DockerAPI) ListImages() ([]types.ImageSummary, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (d *DockerAPI) ListContains() ([]types.Container, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (d *DockerAPI) Attach(container types.Container) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	if container.State != "running" {
		if err := cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	if err := execDocker("attach", container.Names[0]); err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}

func (d *DockerAPI) Run(imageID, containerName string, opt RunOptions) error {
	logging.Debug("container " + containerName + " not found, create new one")

	args := []string{"run", "--name", containerName}
	// moint volumes
	for src, dst := range opt.Mount {
		args = append(args, "-v", src+":"+dst)
	}
	// set env
	for k, v := range opt.Env {
		args = append(args, "-e", k+"="+v)
	}
	// remove container after exit
	if opt.Rm {
		args = append(args, "--rm")
	}
	if opt.Detach {
		args = append(args, "-d")
	}
	args = append(args, imageID)
	args = append(args, opt.Command...)

	var runner func(args ...string) error
	if opt.Fork {
		runner = runDocker
	} else {
		runner = execDocker
	}

	if err := runner(args...); err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}

func (d *DockerAPI) Exec(container string, cmd []string, opt ExecOptions) error {
	args := []string{"exec", "-i"}
	if opt.Tty {
		args = append(args, "-t")
	}
	if opt.workDir != "" {
		args = append(args, "-w", opt.workDir)
	}
	args = append(args, container)
	args = append(args, cmd...)

	var runner func(args ...string) error
	if opt.Fork {
		runner = runDocker
	} else {
		runner = execDocker
	}
	if err := runner(args...); err != nil {
		return fmt.Errorf("failed to create shell: %w", err)
	}

	return nil
}

func runDocker(args ...string) error {
	logging.Debug("run docker with docker: ", docker_exec)
	logging.Debug("run docker with args: ", args)
	cmd := exec.Command(docker_exec, args...)

	// must allcate a pty or some progrrams will not output anything
	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer f.Close()

	go func() { io.Copy(os.Stdout, f) }()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func execDocker(args ...string) error {
	logging.Debug("exec docker with docker: ", docker_exec)
	logging.Debug("exec docker with args: ", args)
	args = append([]string{"docker"}, args...)
	if err := syscall.Exec(docker_exec, args, []string{}); err != nil {
		return err
	}

	return nil
}
