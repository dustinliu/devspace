package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	appImageURL = "https://github.com/neovim/neovim/releases/download/stable/nvim.appimage"
	nvimExec    = "nvim"
)

type BaseEnv interface {
	Setup() error
}

type BaseEnvImpl struct {
	repoDir string
	binDir  string
	nvim    string
}

type rdir string

func newBaseEnv(r rdir) *BaseEnvImpl {
	binDir := filepath.Join(string(r), "bin")
	nvim := filepath.Join(binDir, nvimExec)
	return &BaseEnvImpl{
		repoDir: string(r),
		binDir:  binDir,
		nvim:    nvim,
	}
}

func (e *BaseEnvImpl) Setup() error {
	if err := e.buildDirctory(); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}

	if !isPathExisting(e.nvim) {
		if err := e.downloadNeoVim(); err != nil {
			return fmt.Errorf("failed to download neovim: %w", err)
		}
	}

	return nil
}

func (e *BaseEnvImpl) buildDirctory() error {
	if !isPathExisting(e.repoDir) {
		if err := os.MkdirAll(e.repoDir, 0740); err != nil {
			return fmt.Errorf("create %s: %w", e.repoDir, err)
		}
	}

	if !isPathExisting(e.binDir) {
		if err := os.MkdirAll(e.binDir, 0740); err != nil {
			return fmt.Errorf("create %s: %w", e.binDir, err)
		}
	}

	return nil
}

func (e *BaseEnvImpl) downloadNeoVim() error {
	ch := make(chan error)
	Print("Downloading neovim")
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
			Print(".")
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

	Debug("Downloading ", fullURLFile)
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
