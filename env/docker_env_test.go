package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DockerEnvTestSuite struct {
	*EnvTestSuite
}

func (suite *DockerEnvTestSuite) TestCommand() {
	env := &DockerEnv{
		dotfileScripts: []string{"bootstrap"},
	}

	// bootstrap not exist
	p := filepath.Clean(filepath.Join(env.DotfileDir(), "/.[a-zA-Z0-9]*"))
	suite.Equal([]string{"/bin/sh", "-c", "ln -s " + p + " $HOME/"},
		env.BootstrapCommand(suite.DotfilesDir))

	// bootstrap exist
	b := filepath.Join(suite.DotfilesDir, "bootstrap")
	file, err := os.OpenFile(b, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		suite.Fail("create bootstrap script failed")
	}
	file.Close()
	suite.Equal([]string{"/bin/sh", filepath.Join(env.DotfileDir(), "bootstrap")},
		env.BootstrapCommand(suite.DotfilesDir))
}

func TestDockerEnv(t *testing.T) {
	suite.Run(t, &DockerEnvTestSuite{EnvTestSuite: NewEnvTestSuite()})
}
