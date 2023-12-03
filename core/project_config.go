package core

import (
	"fmt"
	"path/filepath"

	"github.com/dustinliu/devspace/env"
	"github.com/dustinliu/devspace/logging"
	"github.com/spf13/viper"
)

const (
	ImageKey             = "image"
	DockerFielKey        = "dockerfile"
	PostCreateCommandKey = "postCreateCommand"
	ShellKey             = "shell"
	DotfilesKey          = "dotfiles"
)

// TODO: add root pattern to config
// TODO: validate config, image and dockerfile conflict
var (
	confName = "config"
	confType = "json"
)

func newViper() *viper.Viper {
	return viper.New()
}

type ProjectConfig interface {
	Image() string
	Dockerfile() string
	PostCreateCommand() string
	Shell() string
	Dotfiles() string
}

type ProjectConfigImpl struct {
	viper *viper.Viper
}

// TODO: root might not be the project dir
func newProjectConfig(v *viper.Viper, projectDir string) *ProjectConfigImpl {
	// setup viper
	confDir := filepath.Join(string(projectDir), env.SpaceName)
	v.AddConfigPath(confDir)
	v.SetConfigName(confName)
	v.SetConfigType(confType)
	v.SetDefault(ShellKey, "/bin/zsh")
	err := v.ReadInConfig()
	if err != nil {
		logging.Fatal(fmt.Errorf("error reading config file: %w", err))
	}

	return &ProjectConfigImpl{
		viper: v,
	}
}

func (c *ProjectConfigImpl) Image() string {
	return c.viper.GetString(ImageKey)
}

func (c *ProjectConfigImpl) Dockerfile() string {
	return c.viper.GetString(DockerFielKey)
}

func (c *ProjectConfigImpl) PostCreateCommand() string {
	return c.viper.GetString(PostCreateCommandKey)
}

func (c *ProjectConfigImpl) Shell() string {
	return c.viper.GetString(ShellKey)
}

func (c *ProjectConfigImpl) Dotfiles() string {
	return c.viper.GetString(DotfilesKey)
}
