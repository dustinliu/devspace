package env

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dustinliu/devspace/logging"
)

const (
	appImageURL = "https://github.com/neovim/neovim/releases/download/stable/nvim.appimage"
	nvimExec    = "nvim"
)

type BaseEnv interface {
	Setup() error
	RepoDir() string
}

type BaseEnvAppImage struct {
	repoDir string
	binDir  string
	nvim    string
}

type rdir string

func newBaseEnv(r rdir) *BaseEnvAppImage {
	binDir := filepath.Join(string(r), "bin")
	nvim := filepath.Join(binDir, nvimExec)
	return &BaseEnvAppImage{
		repoDir: string(r),
		binDir:  binDir,
		nvim:    nvim,
	}
}

func (e *BaseEnvAppImage) Setup() error {
	if err := e.buildDirctory(); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	if !IsPathExisting(e.nvim) {
		if err := e.downloadNeoVim(); err != nil {
			return fmt.Errorf("failed to download neovim: %w", err)
		}
	}

	return nil
}

func (e *BaseEnvAppImage) RepoDir() string {
	return e.repoDir
}

func (e *BaseEnvAppImage) buildDirctory() error {
	if !IsPathExisting(e.repoDir) {
		if err := os.MkdirAll(e.repoDir, 0740); err != nil {
			return fmt.Errorf("create %s: %w", e.repoDir, err)
		}
	}

	if !IsPathExisting(e.binDir) {
		if err := os.MkdirAll(e.binDir, 0740); err != nil {
			return fmt.Errorf("create %s: %w", e.binDir, err)
		}
	}

	return nil
}

func (e *BaseEnvAppImage) downloadNeoVim() error {
	ch := make(chan error)
	fmt.Print("Downloading neovim")
	go download(appImageURL, e.nvim, ch)
LOOP:
	for {
		select {
		case err := <-ch:
			if err != nil {
				return fmt.Errorf("neovim download failed: %w", err)
			}
			if err := os.Chmod(e.nvim, 0755); err != nil {
				return fmt.Errorf("nvim chmod failed: %w", err)
			}
			break LOOP
		default:
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func download(fullURLFile, dest string, ch chan (error)) {
	file, err := os.Create(dest)
	if err != nil {
		ch <- err
	}
	defer file.Close()

	client := http.Client{
		CheckRedirect: func(r *http.Request, _ []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	logging.Debug("Downloading ", fullURLFile)
	resp, err := client.Get(fullURLFile)
	if err != nil {
		ch <- err
	}
	defer resp.Body.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		ch <- err
	}

	ch <- nil
}
