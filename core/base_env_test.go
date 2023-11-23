package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BaseEnvTestSuite struct {
	suite.Suite
	repoDir string
}

func purgeDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}

	return nil
}

func (s *BaseEnvTestSuite) SetupTest() {
	purgeDir(s.repoDir)
}

func (s *BaseEnvTestSuite) TestEnsure() {
	s.False(isPathExisting(s.repoDir))
	env := newBaseEnv(rdir(s.repoDir))

	s.Equal(env.repoDir, s.repoDir)
	s.Equal(env.binDir, filepath.Join(s.repoDir, "bin"))
	s.NoError(env.Ensure())
	s.True(isPathExisting(env.binDir))
	s.True(isPathExisting(env.nvim))
}

func TestBaseEnv(t *testing.T) {
	var r = filepath.Join(os.TempDir(), "devspace")
	s := &BaseEnvTestSuite{repoDir: r}
	suite.Run(t, s)
}
