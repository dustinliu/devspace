package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func isPathExisting(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return false
}

func md5sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func execCmd(args ...string) error {
	cmd := exec.Command(docker_exec, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() { io.Copy(os.Stdout, stdout) }()
	go func() { io.Copy(os.Stderr, stderr) }()

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
