package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EncryptTestSuite struct {
	suite.Suite
	file1 string
	file2 string
}

func (suite *EncryptTestSuite) SetupSuite() {
	suite.file1 = filepath.Join(os.TempDir(), "file1")
	suite.file2 = filepath.Join(os.TempDir(), "file2")
}

func (suite *EncryptTestSuite) TearDownSuite() {
	os.Remove(suite.file1)
	os.Remove(suite.file2)
}

func (suite *EncryptTestSuite) TestEncrypt() {
	suite.Require().Nil(os.WriteFile(suite.file1, []byte("Hello"), 0644))
	suite.Require().Nil(os.WriteFile(suite.file2, []byte("World"), 0644))

	m1, err := md5sum("test", suite.file1, suite.file2)
	suite.Nil(err)

	suite.Require().Nil(os.WriteFile(suite.file2, []byte("World11"), 0644))
	m2, err := md5sum("test", suite.file1, suite.file2)
	suite.Nil(err)

	suite.NotEqual(m1, m2)
}

func (suite *EncryptTestSuite) TestEncryptPrefix() {
	suite.Require().Nil(os.WriteFile(suite.file1, []byte("Hello"), 0644))
	suite.Require().Nil(os.WriteFile(suite.file2, []byte("World"), 0644))

	m1, err := md5sum("test", suite.file1, suite.file2)
	suite.Nil(err)

	m2, err := md5sum("test1", suite.file1, suite.file2)
	suite.Nil(err)

	suite.NotEqual(m1, m2)
}

func TestEncrypt(t *testing.T) {
	suite.Run(t, new(EncryptTestSuite))
}
