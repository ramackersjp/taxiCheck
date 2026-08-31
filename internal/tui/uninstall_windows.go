//go:build windows

package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/ramackersjp/taxiCheck/internal/wininstall"
)

const uninstallRegKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\TaxiCheck`

func uninstallAppImpl() error {
	dir := wininstall.UserInstallDir()
	_ = removeUserPath(dir)
	_ = removeStartMenuShortcut()
	_ = registry.DeleteKey(registry.CURRENT_USER, uninstallRegKey)

	if home, err := osUserHome(); err == nil && home != "" {
		_ = os.RemoveAll(filepath.Join(home, wininstall.ConfigDirName))
	}

	if dir != "" {
		for _, name := range []string{
			wininstall.BinName,
			".env",
			".env.example",
			"LICENSE",
			wininstall.LauncherName,
			wininstall.UninstallName,
		} {
			_ = os.Remove(filepath.Join(dir, name))
		}
		_ = os.RemoveAll(dir)
		if _, err := os.Stat(dir); err == nil {
			_ = scheduleRemoveDir(dir)
		}
	}
	if running := runningExecutable(); running != "" {
		_ = os.Remove(running)
	}
	return nil
}

func startMenuDir() string {
	if d := os.Getenv("APPDATA"); d != "" {
		return filepath.Join(d, "Microsoft", "Windows", "Start Menu", "Programs")
	}
	home, _ := osUserHome()
	return filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs")
}

func removeStartMenuShortcut() error {
	dir := startMenuDir()
	_ = os.Remove(filepath.Join(dir, wininstall.AppName+".lnk"))
	_ = os.Remove(filepath.Join(dir, wininstall.LauncherName))
	return nil
}

func removeUserPath(dir string) error {
	if dir == "" {
		return nil
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	cur, _, err := k.GetStringValue("Path")
	if err == registry.ErrNotExist {
		return nil
	}
	if err != nil {
		return err
	}
	next := wininstall.RemovePath(cur, dir)
	if next == cur {
		return nil
	}
	if err := k.SetExpandStringValue("Path", next); err != nil {
		if err := k.SetStringValue("Path", next); err != nil {
			return err
		}
	}
	notifyEnvChange()
	return nil
}

func notifyEnvChange() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("SendMessageTimeoutW")
	env, err := windows.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}
	var result uintptr
	_, _, _ = proc.Call(
		uintptr(0xffff),
		uintptr(0x001A),
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(0x0002),
		uintptr(5000),
		uintptr(unsafe.Pointer(&result)),
	)
}

func scheduleRemoveDir(dir string) error {
	script := filepath.Join(os.TempDir(), "taxicheck-uninstall.cmd")
	body := "@echo off\r\n" +
		"ping 127.0.0.1 -n 3 >nul\r\n" +
		"rmdir /s /q \"" + dir + "\"\r\n" +
		"del \"%~f0\"\r\n"
	if err := os.WriteFile(script, []byte(body), 0644); err != nil {
		return err
	}
	cmd := exec.Command("cmd.exe", "/C", "start", "/min", "", script)
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NO_WINDOW | windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd.Start()
}
