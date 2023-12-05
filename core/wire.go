//go:build wireinject
// +build wireinject

package core

import (
	"github.com/dustinliu/devspace/env"
	"github.com/google/wire"
)

func initProject(p string) *Project {
	wire.Build(
		newProjectInternal,
		newProjectConfig,
		env.NewDockerEnv,
		newViper,
		newDocker,
	)

	return &Project{}
}
