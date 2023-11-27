//go:build wireinject
// +build wireinject

package core

import (
	"github.com/google/wire"
)

func initProject(r rdir, p pdir) *Project {
	wire.Build(
		newProjectInternal,
		newProjectConfig,
		wire.Bind(new(ProjectConfig), new(*ProjectConfigImpl)),
		newBaseEnv,
		wire.Bind(new(BaseEnv), new(*BaseEnvImpl)),
		newViper,
		newDocker,
		wire.Bind(new(Docker), new(*DockerAPI)),
	)

	return &Project{}
}
