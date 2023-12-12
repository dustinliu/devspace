package core

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

const (
	testPrjName = "test-project"
)

type CoreTestSuite struct {
	*suite.Suite
	TestDir     string
	ProjectDir  string
	ProjectName string
	ShareSpace  string
	DotfilesDir string
}

// remove test space directory then create test project and share space
func (s *CoreTestSuite) SetupTest() {
	s.Require().Nil(os.RemoveAll(s.TestDir))
	s.Require().Nil(os.MkdirAll(s.ProjectDir, 0755))
}

func NewCoreTestSuite() CoreTestSuite {
	r := filepath.Join(os.TempDir(), "devspace")
	return CoreTestSuite{
		Suite:       new(suite.Suite),
		TestDir:     r,
		ProjectDir:  filepath.Join(r, testPrjName),
		ProjectName: testPrjName,
	}
}
