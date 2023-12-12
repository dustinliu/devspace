package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dustinliu/devspace/env"
	"github.com/stretchr/testify/suite"
)

const (
	test_config = `image = "node:latest"
dockerfile = "Dockerfile"
postCreateCommand = ["npm", "install"]
dotfiles = "~/ini"
rootPattern = [".git", "go.mod"]
`
)

type ProjectConfigTestSuite struct {
	CoreTestSuite
}

func (s *ProjectConfigTestSuite) SetupTest() {
	s.CoreTestSuite.SetupTest()
}

func (s *ProjectConfigTestSuite) TestLoadConfig() {
	confDir := filepath.Join(s.ProjectDir, env.SpaceName)
	s.Require().Nil(os.MkdirAll(confDir, 0755))
	s.Require().Nil(os.WriteFile(filepath.Join(confDir, "config"), []byte(test_config), 0644))

	config := newProjectConfig(s.ProjectDir)
	s.Equal("node:latest", config.Image())
	s.Equal("Dockerfile", config.Dockerfile())
	s.Equal([]string{"npm", "install"}, config.PostCreateCommand())
	s.Equal("~/ini", config.Dotfiles())
	s.Equal([]string{".git", "go.mod"}, config.RootPattern())
}

func TestProjectConfig(t *testing.T) {
	suite.Run(t, &ProjectConfigTestSuite{NewCoreTestSuite()})
}
