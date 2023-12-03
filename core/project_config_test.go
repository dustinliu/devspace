package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

// TODO: remove the dir
type ProjectConfigTestSuite struct {
	DevspaceTestSuite
	TestDir    string
	projectDir string
}

func (s *ProjectConfigTestSuite) TestLoadConfig() {
	v := viper.New()
	v.SetConfigName("simple.hcl")
	v.SetConfigType("hcl")

	projectDir, err := findProjectRoot(".")
	if err != nil {
		s.Fail("failed to find project root")
		panic(err)
	}
	projectDir = filepath.Join(projectDir, "tests")
	config := newProjectConfig(v, (projectDir))

	s.Equal("node:latest", config.Image())
	s.Equal("Dockerfile", config.Dockerfile())
	s.Equal("npm install", config.PostCreateCommand())
	s.Equal("~/ini", config.Dotfiles())
}

func (s *ProjectConfigTestSuite) SetupTest() {
	os.RemoveAll(s.TestDir)
	if err := os.MkdirAll(s.projectDir, 0755); err != nil {
		s.Fail("failed to create test project directory")
	}
}

// TODO: wrong suite to new
func TestProjectConfig(t *testing.T) {
	suite.Run(t, NewDevspaceTestSuite())
}
