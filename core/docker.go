package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/dustinliu/devspace/env"
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
	Detach  bool
	Mount   map[string]string
	Command []string
	Env     map[string]string
	WorkDir string
	Labels  map[string]string
}

type ExecOptions struct {
	WorkDir string
	User    string
}

func newDocker() *Docker {
	return &Docker{}
}

type Docker struct{}

func (d *Docker) BuildImage(opts types.ImageBuildOptions, path string) error {
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
		bytes := scanner.Bytes()
		logging.Debug("build response: ", string(bytes))
		if err := json.Unmarshal(bytes, &r); err != nil {
			return err
		}
		if r.Stream != "" {
			fmt.Print(r.Stream)
		}
	}

	return nil
}

func (d *Docker) ListImages() ([]types.ImageSummary, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	filter := filters.NewArgs(filters.Arg("Labels", env.SpaceName))
	images, err := cli.ImageList(ctx, types.ImageListOptions{Filters: filter})
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (d *Docker) ListContains(opts types.ContainerListOptions) ([]types.Container, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	logging.Debug("number of container: ", len(containers))
	logging.Debugf("containers: %v", containers)
	return containers, nil
}

func (d *Docker) StartContainer(id string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

func (d *Docker) StopContainer(id string, timeout int) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	if timeout == 0 {
		timeout = 10
	}
	opts := container.StopOptions{
		Timeout: &timeout,
	}
	if err := cli.ContainerStop(ctx, id, opts); err != nil {
		return err
	}

	return nil
}

func (d *Docker) Run(imageID, containerName string, opt RunOptions) error {
	logging.Debug("container " + containerName + " not found, create new one")

	args := []string{"run", "--name", containerName}
	// set labels
	for k, v := range opt.Labels {
		args = append(args, "-l", k+"="+v)
	}
	// mount volumes
	for src, dst := range opt.Mount {
		args = append(args, "-v", src+":"+dst)
	}
	// set env
	for k, v := range opt.Env {
		args = append(args, "-e", k+"="+v)
	}
	if opt.Detach {
		args = append(args, "-d")
	}
	args = append(args, imageID)
	args = append(args, opt.Command...)

	if err := runDocker(args...); err != nil {
		return fmt.Errorf("failed to run shell: %w", err)
	}

	return nil
}

func (d *Docker) Exec(container string, cmd []string, opt ExecOptions) error {
	args := []string{"exec", "-it"}
	if opt.WorkDir != "" {
		args = append(args, "-w", opt.WorkDir)
	}
	if opt.User != "" {
		args = append(args, "-u", opt.User)
	}
	args = append(args, container)
	args = append(args, cmd...)

	if err := runDocker(args...); err != nil {
		return fmt.Errorf("failed to create shell: %w", err)
	}

	return nil
}

func runDocker(args ...string) error {
	logging.Debug("run docker with docker: ", docker_exec)
	logging.Debug("run docker with args: ", args)
	cmd := exec.Command(docker_exec, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.Stdin = nil

	return nil
}

func NormalizeContainerName(name string) string {
	return "/" + name
}
