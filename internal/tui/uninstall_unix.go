//go:build unix

package tui

import (
	"os"
	"os/exec"
	"path/filepath"
)

func uninstallAppImpl() error {
	home, _ := osUserHome()
	if home != "" {
		_ = os.Remove(filepath.Join(home, ".local", "bin", "taxiprijs"))
		_ = os.Remove(filepath.Join(home, ".local", "share", "applications", "taxiprijs.desktop"))
		_ = os.RemoveAll(filepath.Join(home, ".taxiprijs"))
		_ = os.RemoveAll(filepath.Join(home, ".config", "omarchy", "plugins", "jp.taxiprijs"))
		removeOmarchyBarEntry(home)
	}
	// Passwordless only: a TUI cannot prompt for sudo. User files are enough
	// for a working uninstall; leftover system files are ignored.
	_ = exec.Command("sudo", "-n", "rm", "-f",
		"/usr/local/bin/taxiprijs",
		"/usr/share/applications/taxiprijs.desktop",
		"/usr/local/share/man/man1/taxiprijs.1",
	).Run()
	if running := runningExecutable(); running != "" {
		_ = os.Remove(running)
	}
	return nil
}

func removeOmarchyBarEntry(home string) {
	cfg := filepath.Join(home, ".config", "omarchy", "shell.json")
	if _, err := os.Stat(cfg); err != nil {
		return
	}
	jq, err := exec.LookPath("jq")
	if err != nil {
		return
	}
	cmd := exec.Command(jq,
		`(.bar.layout.center // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.left // []) |= map(select(.id != "jp.taxiprijs")) | (.bar.layout.right // []) |= map(select(.id != "jp.taxiprijs"))`,
		cfg,
	)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	tmp := cfg + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return
	}
	_ = os.Rename(tmp, cfg)
}
