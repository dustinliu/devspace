package core

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

const (
	testPrjName = "test-project"
)

type DevspaceTestSuite struct {
	*suite.Suite
	TestDir     string
	ProjectDir  string
	ProjectName string
	ShareSpace  string
	DotfilesDir string
}

// remove test space directory then create test project and share space
func (s *DevspaceTestSuite) SetupTest() {
	s.Require().NoError(os.RemoveAll(s.TestDir))
	s.Require().NoError(os.MkdirAll(s.ProjectDir, 0755))
}

func NewDevspaceTestSuite() *DevspaceTestSuite {
	r := filepath.Join(os.TempDir(), "devspace")
	return &DevspaceTestSuite{
		Suite:       new(suite.Suite),
		TestDir:     r,
		ProjectDir:  filepath.Join(r, testPrjName),
		ProjectName: testPrjName,
	}
}
