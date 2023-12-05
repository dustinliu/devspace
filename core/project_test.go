package core

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	CoreTestSuite
}

func (s *ProjectTestSuite) TestNewProject() {
	project := &Project{}

	s.Equal(s.ProjectDir, project.projectDir)
	s.Equal(s.ProjectName, project.projectName)
}

func TestProject(t *testing.T) {
	suite.Run(t, NewCoreTestSuite())
}
