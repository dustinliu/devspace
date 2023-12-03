package env

import (
	"path/filepath"

	"github.com/dustinliu/devspace/logging"
)

type DockerEnv struct {
	workSpace      string
	dotfileScripts []string
}

// TODO: read script from config
func NewDockerEnv() *DockerEnv {
	env := &DockerEnv{}
	env.workSpace = WorkSpace
	env.dotfileScripts = []string{"bootstrap"}
	return env
}

func (e *DockerEnv) WorkSpace() string {
	return e.workSpace
}

func (e *DockerEnv) ShareSpace() string {
	return filepath.Join(e.workSpace, SpaceName)
}

func (e *DockerEnv) DotfileDir() string {
	return filepath.Join(e.ShareSpace(), "dotfiles")
}

func (e *DockerEnv) SetupCommand(dotfileDir string) []string {
	for _, script := range e.dotfileScripts {
		s := filepath.Join(dotfileDir, script)
		if IsFileExisting(s) {
			logging.Debug("Found dotfile script: ", s)
			return []string{"/bin/sh", filepath.Join(e.DotfileDir(), script)}
		}
	}
	logging.Debug("No dotfile script found, use default")
	p := filepath.Clean(filepath.Join(e.DotfileDir(), "/.*"))
	return []string{"/bin/sh", "-c", "ln", "-s", p, "$HOME/"}
}
