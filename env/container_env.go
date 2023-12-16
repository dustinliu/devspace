package env

import (
	"path/filepath"

	"github.com/dustinliu/devspace/logging"
)

type ContainerEnv struct {
	workSpace      string
	dotfileScripts []string
}

func NewContainerEnv() *ContainerEnv {
	env := &ContainerEnv{}
	env.workSpace = WorkSpace
	env.dotfileScripts = []string{"bootstrap"}
	return env
}

func (e *ContainerEnv) WorkSpace() string {
	return e.workSpace
}

func (e *ContainerEnv) ShareSpace() string {
	return filepath.Join(e.workSpace, SpaceName)
}

func (e *ContainerEnv) DotfileDir() string {
	return filepath.Join(e.ShareSpace(), "dotfiles")
}

func (e *ContainerEnv) BootstrapCommand(dotfileDir string) []string {
	for _, script := range e.dotfileScripts {
		s := filepath.Join(dotfileDir, script)
		if IsFileExisting(s) {
			logging.Debug("Found dotfile script: ", s)
			return []string{"/bin/sh", filepath.Join(e.DotfileDir(), script)}
		}
	}
	logging.Debug("No dotfile script found, use default")
	p := e.DotfileDir() + "/.[a-zA-Z0-9]*"
	return []string{"/bin/sh", "-c", "ln -s " + p + " $HOME/"}
}
