//go:build unix

package tui

import (
	"fmt"
	"os"
	"syscall"
)

func execReplacedProcessImpl(bin string) error {
	if bin == "" {
		return fmt.Errorf("empty binary path")
	}
	return syscall.Exec(bin, append([]string{bin}, os.Args[1:]...), os.Environ())
}
