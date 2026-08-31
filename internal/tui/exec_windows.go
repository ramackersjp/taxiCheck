//go:build windows

package tui

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func execReplacedProcessImpl(bin string) error {
	if bin == "" {
		return os.ErrNotExist
	}
	cmd := exec.Command(bin)
	cmd.Dir = filepath.Dir(bin)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	return cmd.Start()
}
