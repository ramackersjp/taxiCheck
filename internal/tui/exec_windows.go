//go:build windows

package tui

import "fmt"

func execReplacedProcessImpl(bin string) error {
	return fmt.Errorf("restart the app to run the new version")
}
