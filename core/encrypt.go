package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func md5sum(prefix string, files ...string) (string, error) {
	h := md5.New()
	if _, err := io.WriteString(h, prefix); err != nil {
		return "", err
	}
	for _, f := range files {
		f, err := os.Open(f)
		if err != nil {
			return "", err
		}
		defer f.Close()

		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
