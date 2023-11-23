package core

import (
	"bufio"
	"io"
	"os"
	"os/exec"
)

var docker_exec string

type DockerRunner interface {
	Build() error
}

func init() {
	var err error
	if docker_exec, err = exec.LookPath("docker"); err != nil {
		Fatal("docker not found")
	}
}

type DockerRunnerImpl struct {
}

func (d *DockerRunnerImpl) Build(tag, dockerfile, path string) error {
	return execCmd("build", "-t", tag, "-f", dockerfile, path)
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

	go pipe(stdout, os.Stdout)
	go pipe(stderr, os.Stderr)

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func pipe(pipe io.ReadCloser, f *os.File) {
	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString('\n')
	for err == nil {
		f.WriteString(line)
		line, err = reader.ReadString('\n')
	}
}
