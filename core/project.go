package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/dustinliu/devspace/env"
	"github.com/dustinliu/devspace/logging"
	homedir "github.com/mitchellh/go-homedir"
)

var rootMaxDepth = 4

type Project struct {
	config    *ProjectConfig
	dockerEnv *env.DockerEnv
	docker    *Docker

	projectDir     string
	projectConfDir string
	projectName    string
}

func NewProject() (*Project, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	if !env.IsPathExisting(filepath.Join(currentDir, env.SpaceName)) {
		return nil, errors.New(".devspace directory not found, make sure you are in project root")
	}

	return initProject(currentDir), nil
}

func (p *Project) Shell(stop bool) error {
	container, err := p.findContainer()
	if err != nil {
		return fmt.Errorf("failed to list container: %w", err)
	}
	if container == nil {
		if err := p.createContainer(); err != nil {
			return fmt.Errorf("failed to create container: %w", err)
		}
	} else if container.State != "running" {
		if err := p.docker.StartContainer(container.ID); err != nil {
			return fmt.Errorf("failed to start container: %w", err)
		}
	}

	// exec shell
	exec_opt := ExecOptions{
		WorkDir: filepath.Join(p.dockerEnv.WorkSpace(), p.projectName),
		User:    p.config.User(),
	}
	if err := p.docker.Exec(p.containerName(), []string{p.config.Shell()}, exec_opt); err != nil {
		return fmt.Errorf("failed to exec shell: %w", err)
	}

	if !stop {
		var ans string
		fmt.Print("Do you want to stop the container? [y/N]: ")
		fmt.Scanln(&ans)
		if strings.ToLower(ans)[0] == 'y' {
			stop = true
		}
	}

	if stop {
		return p.docker.StopContainer(container.ID, 2)
	}

	return nil
}

func (p *Project) createContainer() error {
	// build image if dockerfile exists
	if p.config.Dockerfile() != "" {
		opts := types.ImageBuildOptions{
			Dockerfile: p.config.Dockerfile(),
			Tags:       []string{p.imageName()},
			Labels: map[string]string{
				env.SpaceName: p.projectName,
			},
		}
		p.docker.BuildImage(opts, p.projectConfDir)
	}
	// create container
	opt := RunOptions{
		Command: []string{"/bin/sleep", "infinity"},
		Env: map[string]string{
			"DEVSPACE":          "true",
			"DEVSPACE_SHARE":    p.dockerEnv.ShareSpace(),
			"DEVSPACE_DOTFILES": p.dockerEnv.DotfileDir(),
		},
		Labels: map[string]string{
			env.SpaceName: p.projectName,
		},
		Mount:   map[string]string{},
		WorkDir: filepath.Join(p.dockerEnv.WorkSpace(), p.projectName),
		Detach:  true,
	}

	dotfiles, err := homedir.Expand(p.config.Dotfiles())
	if err != nil {
		return fmt.Errorf("failed to expand dotfiles path: %w", err)
	}
	if p.config.Dotfiles() != "" {
		opt.Mount[dotfiles] = p.dockerEnv.DotfileDir()
	}
	opt.Mount[p.projectDir] = filepath.Join(p.dockerEnv.WorkSpace(), p.projectName)

	if err := p.docker.Run(p.imageName(), p.containerName(), opt); err != nil {
		return fmt.Errorf("failed to create shell: %w", err)
	}

	// run post create command
	exec_opt := ExecOptions{}
	PostCreateCommand := p.config.PostCreateCommand()
	if len(PostCreateCommand) > 0 {
		exec_opt.WorkDir = filepath.Join(p.dockerEnv.WorkSpace(), p.projectName)
		if err := p.docker.Exec(p.containerName(), PostCreateCommand, exec_opt); err != nil {
			return fmt.Errorf("failed to run dotfiles: %w", err)
		}
	}

	// run dotfiles bootstrap script
	exec_opt.User = p.config.User()
	if err := p.docker.Exec(p.containerName(), p.dockerEnv.BootstrapCommand(dotfiles), exec_opt); err != nil {
		return fmt.Errorf("failed to bootstrap dotfiles: %w", err)
	}

	return nil
}

func (p *Project) findContainer() (*types.Container, error) {
	filter := filters.NewArgs(filters.Arg("name", NormalizeContainerName(p.containerName())))
	opts := types.ContainerListOptions{
		All:     true,
		Filters: filter,
	}
	containers, err := p.docker.ListContains(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find container: %w", err)
	}

	if len(containers) == 0 {
		return nil, nil
	} else if len(containers) > 1 {
		logging.Fatal("found multiple containers with same name")
	}

	return &containers[0], nil
}

func (p *Project) imageName() string {
	if p.config.Dockerfile() != "" {
		return p.projectName + "-" + p.md5("image")
	}

	return p.config.Image()
}

func (p *Project) containerName() string {
	return p.projectName + "-" + p.md5("container")
}

func (p *Project) md5(prefix string) string {
	files := []string{filepath.Join(p.projectConfDir, confName)}
	if p.config.Dockerfile() != "" {
		files = append(files, filepath.Join(p.projectConfDir, p.config.Dockerfile()))
	}
	sum, err := md5sum(prefix, files...)
	if err != nil {
		logging.Fatal(fmt.Errorf("failed to get md5sum of config file: %w", err))
	}
	return sum
}

func newProjectInternal(projectDir string, config *ProjectConfig, dockerEnv *env.DockerEnv, docker *Docker) *Project {
	return &Project{
		config:         config,
		dockerEnv:      dockerEnv,
		projectDir:     projectDir,
		projectConfDir: filepath.Join(string(projectDir), env.SpaceName),
		projectName:    filepath.Base(string(projectDir)),
		docker:         docker,
	}
}
