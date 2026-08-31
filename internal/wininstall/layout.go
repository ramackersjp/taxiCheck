package wininstall

import (
	"os"
	"path/filepath"
)

// UserInstallDir is %LOCALAPPDATA%\TaxiCheck on Windows (user-level install).
func UserInstallDir() string {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, AppName)
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, "AppData", "Local", AppName)
}

func UserInstallBin() string {
	if dir := UserInstallDir(); dir != "" {
		return filepath.Join(dir, BinName)
	}
	return ""
}
