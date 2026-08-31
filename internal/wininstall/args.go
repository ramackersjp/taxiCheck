// Package wininstall holds shared helpers for the Windows install.exe.
package wininstall

import "strings"

const (
	AppName       = "TaxiCheck"
	BinName       = "taxiprijs.exe"
	LauncherName  = "TaxiCheck.cmd"
	UninstallName = "uninstall.exe"
	ConfigDirName = ".taxiprijs"
)

// ParseArgs reads installer flags from os.Args[1:] (or any equivalent slice).
// Silent covers /S, /silent, -s, --quiet. Uninstall covers /uninstall, /u.
func ParseArgs(args []string) (silent, uninstall bool) {
	for _, raw := range args {
		a := strings.TrimSpace(raw)
		a = strings.TrimLeft(a, "-/")
		switch strings.ToLower(a) {
		case "s", "silent", "quiet", "q":
			silent = true
		case "uninstall", "u":
			uninstall = true
		}
	}
	return silent, uninstall
}

// DisplayVersion normalizes a Makefile / ldflags version for UI text.
func DisplayVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "dev"
	}
	if v != "dev" && !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}
