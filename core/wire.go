//go:build wireinject
// +build wireinject

package core

import (
	"github.com/dustinliu/devspace/env"
	"github.com/google/wire"
)

// TODO: clear the dependency graph
func initProject(p string) *Project {
	wire.Build(
		newProjectInternal,
		newProjectConfig,
		wire.Bind(new(ProjectConfig), new(*ProjectConfigImpl)),
		env.NewDockerEnv,
		newViper,
		newDocker,
		wire.Bind(new(Docker), new(*DockerAPI)),
	)

	return &Project{}
}
