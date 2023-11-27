package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	rootPattern  = []string{".git"}
	rootMaxDepth = 4
)

type Project struct {
	config  ProjectConfig
	baseEnv BaseEnv
	docker  Docker

	projectDir  string
	projectName string
}

func NewProject() (*Project, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	projectDir, err := findProjectRoot(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get project directory: %w", err)
	}

	if !isPathExisting(filepath.Join(projectDir, confDirName)) {
		return nil, errors.New(".devspace directory not found")
	}

	repoDir := filepath.Join(xdg.Home, confDirName)
	return initProject(rdir(repoDir), pdir(projectDir)), nil
}

func (p *Project) Build() error {
	tag, err := p.imageName()
	if err != nil {
		return fmt.Errorf("failed to get tag: %w", err)
	}

	path := filepath.Join(p.projectDir, confDirName)
	if err := p.docker.BuildImage(tag, p.config.Dockerfile(), path); err != nil {
		return err
	}

	return nil
}

func (p *Project) Shell() error {
	image := p.config.Image()
	return nil
}

func (p *Project) imageName() (string, error) {
	sum, err := md5sum(filepath.Join(p.projectDir, confDirName, confName))
	if err != nil {
		return "", fmt.Errorf("failed to get md5sum of config file: %w", err)
	}
	return p.projectName + "-" + sum + ":nvim", nil
}

func findProjectRoot(dir string) (string, error) {
	currentDir, err := filepath.Abs(dir)
	for i := 0; i < rootMaxDepth; i++ {
		if err != nil {
			return "", fmt.Errorf("failed to get project dir %w", err)
		}

		for _, pattern := range rootPattern {
			if isPathExisting(filepath.Join(currentDir, pattern)) {
				return currentDir, nil
			}
		}
		currentDir = filepath.Dir(currentDir)
	}
	return "", errors.New("project root not found")
}

type pdir string

func newProjectInternal(projectDir pdir, config ProjectConfig, baseEnv BaseEnv, docker Docker) *Project {
	return &Project{
		config:      config,
		baseEnv:     baseEnv,
		projectDir:  string(projectDir),
		projectName: filepath.Base(string(projectDir)),
		docker:      docker,
	}
}
