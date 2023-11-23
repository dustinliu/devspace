package core

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

const (
	testPrjNmae = "test-project"
)

type DevspaceTestSuite struct {
	suite.Suite
	TestDir     string
	ProjectDir  string
	ProjectName string
}

// remove test project directory then create project directory
func (s *DevspaceTestSuite) SetupTest() {
	Print("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	os.RemoveAll(s.TestDir)
	if err := os.MkdirAll(s.ProjectDir, 0755); err != nil {
		s.Fail("failed to create test project directory")
	}
}

func NewDevspaceTestSuite() *DevspaceTestSuite {
	r := filepath.Join(os.TempDir(), "devspace")
	p := filepath.Join(r, testPrjNmae)
	return &DevspaceTestSuite{
		TestDir:     r,
		ProjectDir:  p,
		ProjectName: testPrjNmae,
	}
}
