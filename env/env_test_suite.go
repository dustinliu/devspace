package env

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

const (
	testPrjName = "test-project"
)

type EnvTestSuite struct {
	*suite.Suite
	TestDir     string
	ShareSpace  string
	DotfilesDir string
}

// remove test space directory then create test project and share space
func (s *EnvTestSuite) SetupTest() {
	s.Require().NoError(os.RemoveAll(s.TestDir))
	s.Require().NoError(os.MkdirAll(s.DotfilesDir, 0755))
}

func NewEnvTestSuite() *EnvTestSuite {
	r := filepath.Join(os.TempDir(), "devspace")
	d := filepath.Join(r, ".devspaces")
	return &EnvTestSuite{
		Suite:       new(suite.Suite),
		TestDir:     r,
		ShareSpace:  d,
		DotfilesDir: filepath.Join(d, "dotfiles"),
	}
}
