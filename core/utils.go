package core

import (
	"os"
)

func isPathExisting(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return false
}
