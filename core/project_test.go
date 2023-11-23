package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	DevspaceTestSuite
}

func (s *ProjectTestSuite) TestFindProjectRoot() {
	// no .git
	root, err := findProjectRoot(s.ProjectDir)
	s.Empty(root)
	s.NotNil(err)

	// with .git
	git := filepath.Join(s.ProjectDir, ".git")
	if err := os.MkdirAll(git, 0755); err != nil {
		s.Fail("failed to create .git directory")
	}

	root, err = findProjectRoot(s.ProjectDir)
	s.Equal(s.ProjectDir, root)
	s.Nil(err)

	// with .git and in subdirectory
	sub := filepath.Join(s.ProjectDir, "sub")
	root, err = findProjectRoot(sub)
	s.Equal(s.ProjectDir, root)
	s.Nil(err)
}

func (s *ProjectTestSuite) TestNewProject() {
	project := &Project{}

	s.Equal(s.ProjectDir, project.projectDir)
	s.Equal(s.ProjectName, project.projectName)
}

func TestProject(t *testing.T) {
	suite.Run(t, NewDevspaceTestSuite())
}
