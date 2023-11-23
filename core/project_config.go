package core

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ImageKey             = "image"
	DockerFielKey        = "dockerfile"
	PostCreateCommandKey = "postCreateCommand"
)

// TODO: add root pattern to config
var (
	confName    = "config"
	confType    = "hcl"
	confDirName = ".devspace"
)

func newViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(confName)
	v.SetConfigType(confType)
	return v
}

type ProjectConfig interface {
	Image() string
	Dockerfile() string
	PostCreateCommand() string
}

type ProjectConfigImpl struct {
	viper *viper.Viper
}

func newProjectConfig(v *viper.Viper, root pdir) *ProjectConfigImpl {
	confDir := filepath.Join(string(root), confDirName)
	v.AddConfigPath(confDir)
	err := v.ReadInConfig()
	if err != nil {
		Fatal(fmt.Errorf("error reading config file: %w", err))
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
