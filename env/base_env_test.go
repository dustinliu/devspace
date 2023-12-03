package env

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
	s.T().Skip("skip base env test")
	s.False(IsPathExisting(s.repoDir))
	env := newBaseEnv(rdir(s.repoDir))

	s.Equal(env.repoDir, s.repoDir)
	s.Equal(env.binDir, filepath.Join(s.repoDir, "bin"))
	s.NoError(env.Setup())
	s.True(IsPathExisting(env.binDir))
	s.True(IsPathExisting(env.nvim))
}

func TestBaseEnv(t *testing.T) {
	r := filepath.Join(os.TempDir(), "devspace")
	s := &BaseEnvTestSuite{repoDir: r}
	suite.Run(t, s)
}
