package core

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
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
	if err := p.baseEnv.Ensure(); err != nil {
		return fmt.Errorf("failed to build base environment: %w", err)
	}

	sum, err := md5sum(filepath.Join(p.projectDir, confDirName, confName))
	if err != nil {
		return fmt.Errorf("failed to get md5sum of config file: %w", err)
	}
	tag := p.projectName + "-" + sum
	Debug("docker image tag: " + tag)

	dockerfile := filepath.Join(p.projectDir, confDirName, p.config.Dockerfile())
	if err := execCmd("build", "-t", tag, "-f", dockerfile, "."); err != nil {
		return err
	}
	return nil
}

type pdir string

func newProjectInternal(projectDir pdir, config ProjectConfig, baseEnv BaseEnv) *Project {
	return &Project{
		config:      config,
		baseEnv:     baseEnv,
		projectDir:  string(projectDir),
		projectName: filepath.Base(string(projectDir)),
	}
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

func md5sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
