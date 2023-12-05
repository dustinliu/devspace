package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ProjectConfigTestSuite struct {
	CoreTestSuite
	TestDir    string
	projectDir string
}

func (s *ProjectConfigTestSuite) TestLoadConfig() {
	currentDir, err := os.Getwd()
	if err != nil {
		s.Fail("failed to get current working directory")
	}
	projectDir := filepath.Join(currentDir, "tests")
	config := newProjectConfig(viper.New(), projectDir)

	s.Equal("node:latest", config.Image())
	s.Equal("Dockerfile", config.Dockerfile())
	s.Equal("npm install", config.PostCreateCommand())
	s.Equal("~/ini", config.Dotfiles())
	s.Equal([]string{".git", "go.mod"}, config.RootPattern())
}

func (s *ProjectConfigTestSuite) SetupTest() {
	os.RemoveAll(s.TestDir)
	if err := os.MkdirAll(s.projectDir, 0755); err != nil {
		s.Fail("failed to create test project directory")
	}
}

func TestProjectConfig(t *testing.T) {
	suite.Run(t, NewCoreTestSuite())
}
