package core

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	CoreTestSuite
}

func (s *ProjectTestSuite) TestNewProject() {
}

func TestProject(t *testing.T) {
	suite.Run(t, &ProjectTestSuite{NewCoreTestSuite()})
}
