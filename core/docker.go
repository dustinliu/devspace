package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

var docker_exec string

type Docker interface {
	ListImages() ([]types.ImageSummary, error)
	BuildImage(tag, dockerfile, path string) error
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
			Print(r.Stream)
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

func (d *DockerAPI) FindImage(id string) (types.ImageSummary, error) {
	images, err := d.ListImages()
	if err != nil {
		return types.ImageSummary{}, err
	}

	for _, image := range images {
		if id == image.ID {
			return image, nil
		}
	}

	return types.ImageSummary{}, fmt.Errorf("image %s not found", id)
}

func (d *DockerAPI) CreateVolume(name string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.VolumeCreate(ctx, volume.CreateOptions{
		Name: name,
	})
	if err != nil {
		return err
	}

	return nil
}
