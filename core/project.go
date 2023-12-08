package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustinliu/devspace/env"
	"github.com/dustinliu/devspace/logging"
	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/exp/slices"
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

func (p *Project) Build() error {
	tag := p.imageName()

	if err := p.docker.BuildImage(tag, p.config.Dockerfile(), p.projectConfDir); err != nil {
		return err
	}

	return nil
}

func (p *Project) Shell() error {
	container, err := p.findContainer()
	if err != nil {
		return fmt.Errorf("failed to list container: %w", err)
	}
	if container != "" {
		exec_opt := ExecOptions{
			Fork:    false,
			Tty:     true,
			workDir: filepath.Join(p.dockerEnv.WorkSpace(), p.projectName),
			User:    p.config.User(),
		}
		return p.docker.Exec(container, []string{p.config.Shell()}, exec_opt)
	}

	// if dockerfile exists, build image
	if p.config.Dockerfile() != "" {
		p.docker.BuildImage(p.imageName(), p.config.Dockerfile(), p.projectConfDir)
	}

	return p.createContainer()
}

func (p *Project) createContainer() error {
	dotfiles, err := homedir.Expand(p.config.Dotfiles())
	if err != nil {
		return fmt.Errorf("failed to expand dotfiles path: %w", err)
	}
	opt := RunOptions{
		Fork:    true,
		Detach:  true,
		Command: []string{"/bin/sleep", "infinity"},
		Env: map[string]string{
			"DEVSPACE":          "true",
			"DEVSPACE_SHARE":    p.dockerEnv.ShareSpace(),
			"DEVSPACE_DOTFILES": p.dockerEnv.DotfileDir(),
		},
		Mount: map[string]string{},
	}
	if p.config.Dotfiles() != "" {
		opt.Mount[dotfiles] = p.dockerEnv.DotfileDir()
	}
	opt.Mount[p.projectDir] = filepath.Join(p.dockerEnv.WorkSpace(), p.projectName)

	// create container
	if err := p.docker.Run(p.imageName(), p.containerName(), opt); err != nil {
		return fmt.Errorf("failed to create shell: %w", err)
	}

	exec_opt := ExecOptions{}
	PostCreateCommand := p.config.PostCreateCommand()
	if len(PostCreateCommand) > 0 {
		exec_opt.Fork = true
		exec_opt.Tty = true
		exec_opt.workDir = filepath.Join(p.dockerEnv.WorkSpace(), p.projectName)
		if err := p.docker.Exec(p.containerName(), PostCreateCommand, exec_opt); err != nil {
			return fmt.Errorf("failed to run dotfiles: %w", err)
		}
	}

	exec_opt.Fork = true
	exec_opt.Tty = true
	exec_opt.User = p.config.User()
	// run dotfiles bootstrap script
	if err := p.docker.Exec(p.containerName(), p.dockerEnv.BootstrapCommand(dotfiles), exec_opt); err != nil {
		return fmt.Errorf("failed to bootstrap dotfiles: %w", err)
	}

	exec_opt.Fork = false
	exec_opt.Tty = true
	exec_opt.workDir = filepath.Join(p.dockerEnv.WorkSpace(), p.projectName)
	exec_opt.User = p.config.User()
	if err := p.docker.Exec(p.containerName(), []string{p.config.Shell()}, exec_opt); err != nil {
		return fmt.Errorf("failed to create shell: %w", err)
	}

	return nil
}

func (p *Project) findContainer() (string, error) {
	containers, err := p.docker.ListContains()
	if err != nil {
		return "", fmt.Errorf("failed to find container: %w", err)
	}
	cname := "/" + p.containerName()
	for _, container := range containers {
		if slices.Contains(container.Names, cname) && container.Image == p.imageName() {
			return container.Names[0], nil
		}
	}
	return "", nil
}

func (p *Project) imageName() string {
	if p.config.Dockerfile() != "" {
		return p.projectName + "-" + p.md5()
	}

	return p.config.Image()
}

func (p *Project) containerName() string {
	return p.projectName + "-" + p.md5()
}

func (p *Project) md5() string {
	sum, err := md5sum(filepath.Join(p.projectConfDir, confName))
	if err != nil {
		logging.Fatal(fmt.Errorf("failed to get md5sum of config file: %w", err))
	}
	return sum
}

func newProjectInternal(projectDir string, config *ProjectConfig, dockerEnv *env.DockerEnv, docker *Docker) *Project {
	return &Project{
		config:         config,
		dockerEnv:      dockerEnv,
		projectDir:     string(projectDir),
		projectConfDir: filepath.Join(string(projectDir), env.SpaceName),
		projectName:    filepath.Base(string(projectDir)),
		docker:         docker,
	}
}
