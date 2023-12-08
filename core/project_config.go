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
	DockerFileKey        = "dockerfile"
	PostCreateCommandKey = "postCreateCommand"
	ShellKey             = "shell"
	DotfilesKey          = "dotfiles"
	RootPatternKey       = "rootPattern"
)

var (
	confName = "config"
	confType = "json"
)

func newViper() *viper.Viper {
	return viper.New()
}

type ProjectConfig struct {
	viper *viper.Viper
}

func newProjectConfig(v *viper.Viper, projectDir string) *ProjectConfig {
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

	return &ProjectConfig{
		viper: v,
	}
}

func (c *ProjectConfig) Image() string {
	return c.viper.GetString(ImageKey)
}

func (c *ProjectConfig) Dockerfile() string {
	return c.viper.GetString(DockerFileKey)
}

func (c *ProjectConfig) PostCreateCommand() []string {
	return c.viper.GetStringSlice(PostCreateCommandKey)
}

func (c *ProjectConfig) Shell() string {
	return c.viper.GetString(ShellKey)
}

func (c *ProjectConfig) Dotfiles() string {
	return c.viper.GetString(DotfilesKey)
}

func (c *ProjectConfig) User() string {
	return c.viper.GetString("user")
}

func (c *ProjectConfig) RootPattern() []string {
	return c.viper.GetStringSlice(RootPatternKey)
}
