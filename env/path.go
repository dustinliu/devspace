package env

import "os"

const (
	WorkSpace      = "/workspace"
	SpaceName      = ".devspace"
	DotFileDirName = "dotfiles"
)

func IsPathExisting(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return false
}

func IsFileExisting(path string) bool {
	if fi, err := os.Stat(path); err == nil {
		return !fi.IsDir()
	} else if os.IsNotExist(err) {
		return false
	}

	return false
}
