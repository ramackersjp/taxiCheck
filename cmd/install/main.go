//go:build windows

// TaxiCheck Windows installer. Built as dist/install.exe by
// `make build-windows-installer`. Flags: /S (silent), /uninstall.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/ramackersjp/taxiCheck/internal/wininstall"
)

// Set at build time: -X main.version=$(VERSION)
var version = "dev"

const uninstallRegKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\TaxiCheck`

func main() {
	in := bufio.NewReader(os.Stdin)
	silent, uninstall := wininstall.ParseArgs(os.Args[1:])
	if self, err := os.Executable(); err == nil {
		if strings.EqualFold(filepath.Base(self), wininstall.UninstallName) {
			uninstall = true
		}
	}

	var err error
	if uninstall {
		err = runUninstall(in, silent)
	} else {
		err = runInstall(in, silent)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		pause(in, silent)
		os.Exit(1)
	}
	pause(in, silent)
}

func runInstall(in *bufio.Reader, silent bool) error {
	ver := wininstall.DisplayVersion(version)
	dir, err := installDir()
	if err != nil {
		return err
	}
	if len(payloadEXE) == 0 {
		return fmt.Errorf("installer payload is empty; rebuild with make build-windows-installer")
	}

	fmt.Printf("%s %s setup\n\n", wininstall.AppName, ver)
	fmt.Printf("Installing to: %s\n", dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create install directory: %w", err)
	}

	exePath := filepath.Join(dir, wininstall.BinName)
	if err := os.WriteFile(exePath, payloadEXE, 0755); err != nil {
		return fmt.Errorf("cannot write %s (is TaxiCheck running?): %w", wininstall.BinName, err)
	}
	fmt.Printf("  %s\n", wininstall.BinName)

	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), payloadLicense, 0644); err != nil {
		return fmt.Errorf("cannot write LICENSE: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), payloadEnv, 0644); err != nil {
		return fmt.Errorf("cannot write .env.example: %w", err)
	}
	envPath := filepath.Join(dir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		if err := os.WriteFile(envPath, payloadEnv, 0644); err != nil {
			return fmt.Errorf("cannot write .env: %w", err)
		}
		fmt.Println("  .env")
	} else {
		fmt.Println("  .env (kept existing)")
	}

	launcher := launcherScript()
	if err := os.WriteFile(filepath.Join(dir, wininstall.LauncherName), []byte(launcher), 0644); err != nil {
		return fmt.Errorf("cannot write launcher: %w", err)
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate installer: %w", err)
	}
	selfBytes, err := os.ReadFile(self)
	if err != nil {
		return fmt.Errorf("cannot read installer: %w", err)
	}
	uninstallPath := filepath.Join(dir, wininstall.UninstallName)
	if err := os.WriteFile(uninstallPath, selfBytes, 0755); err != nil {
		return fmt.Errorf("cannot write uninstaller: %w", err)
	}

	if err := addUserPath(dir); err != nil {
		return fmt.Errorf("cannot update user PATH: %w", err)
	}
	fmt.Println("  user PATH")

	if err := createStartMenuShortcut(dir, exePath); err != nil {
		fmt.Printf("  Start Menu shortcut skipped: %v\n", err)
	} else {
		fmt.Println("  Start Menu shortcut")
	}

	sizeKB := uint64(len(payloadEXE)+len(payloadEnv)+len(payloadLicense)+len(selfBytes)) / 1024
	if err := writeUninstallReg(dir, uninstallPath, ver, sizeKB); err != nil {
		fmt.Printf("  Add/Remove Programs entry skipped: %v\n", err)
	} else {
		fmt.Println("  Add/Remove Programs")
	}

	fmt.Printf("\nInstallation complete.\n\n")
	fmt.Printf("Open TaxiCheck from the Start Menu, or open a new terminal and run:\n")
	fmt.Printf("  taxiprijs\n")

	if silent {
		return nil
	}
	fmt.Print("\nStart TaxiCheck now? [Y/n] ")
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "" || line == "y" || line == "yes" {
		if err := launchApp(dir, exePath); err != nil {
			fmt.Printf("Could not start TaxiCheck: %v\n", err)
		}
	}
	return nil
}

func runUninstall(in *bufio.Reader, silent bool) error {
	dir, err := installDir()
	if err != nil {
		return err
	}
	fmt.Printf("Uninstall %s from:\n  %s\n\n", wininstall.AppName, dir)
	if !silent {
		fmt.Print("This also removes TaxiCheck settings (.taxiprijs) from your user profile.\nType y to confirm: ")
		line, _ := in.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line != "y" && line != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	_ = removeUserPath(dir)
	_ = removeStartMenuShortcut()
	_ = deleteUninstallReg()

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		_ = os.RemoveAll(filepath.Join(home, wininstall.ConfigDirName))
	}

	for _, name := range []string{
		wininstall.BinName,
		".env",
		".env.example",
		"LICENSE",
		wininstall.LauncherName,
	} {
		_ = os.Remove(filepath.Join(dir, name))
	}

	_ = os.Remove(filepath.Join(dir, wininstall.UninstallName))
	_ = os.RemoveAll(dir)
	if _, err := os.Stat(dir); err == nil {
		_ = scheduleRemoveDir(dir)
	}

	fmt.Println("TaxiCheck has been uninstalled.")
	return nil
}

func installDir() (string, error) {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, wininstall.AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, "AppData", "Local", wininstall.AppName), nil
}

func launcherScript() string {
	return "@echo off\r\n" +
		"title TaxiCheck\r\n" +
		"cd /d \"%~dp0\"\r\n" +
		"taxiprijs.exe\r\n" +
		"if errorlevel 1 pause\r\n"
}

func launchApp(dir, exePath string) error {
	cmd := exec.Command("cmd.exe", "/K", exePath)
	cmd.Dir = dir
	cmd.SysProcAttr = &windows.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE}
	return cmd.Start()
}

func startMenuDir() string {
	if d := os.Getenv("APPDATA"); d != "" {
		return filepath.Join(d, "Microsoft", "Windows", "Start Menu", "Programs")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs")
}

func startMenuLnk() string {
	return filepath.Join(startMenuDir(), wininstall.AppName+".lnk")
}

func startMenuCmd() string {
	return filepath.Join(startMenuDir(), wininstall.LauncherName)
}

func createStartMenuShortcut(dir, exePath string) error {
	if err := os.MkdirAll(startMenuDir(), 0755); err != nil {
		return err
	}
	if err := createLnk(dir, exePath); err == nil {
		if _, err := os.Stat(startMenuLnk()); err == nil {
			return nil
		}
	}
	return copyFile(filepath.Join(dir, wininstall.LauncherName), startMenuCmd())
}

func createLnk(dir, exePath string) error {
	comspec := os.Getenv("ComSpec")
	if comspec == "" {
		root := os.Getenv("SystemRoot")
		if root == "" {
			root = `C:\Windows`
		}
		comspec = filepath.Join(root, "System32", "cmd.exe")
	}
	ps := fmt.Sprintf(
		`$s = (New-Object -ComObject WScript.Shell).CreateShortcut(%s); $s.TargetPath = %s; $s.Arguments = %s; $s.WorkingDirectory = %s; $s.WindowStyle = 1; $s.Description = %s; $s.Save()`,
		psQuote(startMenuLnk()),
		psQuote(comspec),
		psQuote("/K "+quoteArg(exePath)),
		psQuote(dir),
		psQuote("Dutch taxi fare calculator"),
	)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", ps)
	cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_NO_WINDOW}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func removeStartMenuShortcut() error {
	_ = os.Remove(startMenuLnk())
	_ = os.Remove(startMenuCmd())
	return nil
}

func quoteArg(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\"") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func addUserPath(dir string) error {
	cur, err := readUserPath()
	if err != nil {
		return err
	}
	next := wininstall.AppendPath(cur, dir)
	if next == cur {
		return nil
	}
	return writeUserPath(next)
}

func removeUserPath(dir string) error {
	cur, err := readUserPath()
	if err != nil {
		return err
	}
	next := wininstall.RemovePath(cur, dir)
	if next == cur {
		return nil
	}
	return writeUserPath(next)
}

func readUserPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	s, _, err := k.GetStringValue("Path")
	if err == registry.ErrNotExist {
		return "", nil
	}
	return s, err
}

func writeUserPath(value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.SetExpandStringValue("Path", value); err != nil {
		if err := k.SetStringValue("Path", value); err != nil {
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
		uintptr(0xffff), // HWND_BROADCAST
		uintptr(0x001A), // WM_SETTINGCHANGE
		0,
		uintptr(unsafe.Pointer(env)),
		uintptr(0x0002), // SMTO_ABORTIFHUNG
		uintptr(5000),
		uintptr(unsafe.Pointer(&result)),
	)
}

func writeUninstallReg(dir, uninstallPath, ver string, sizeKB uint64) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, uninstallRegKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	_ = k.SetStringValue("DisplayName", wininstall.AppName)
	_ = k.SetStringValue("DisplayVersion", ver)
	_ = k.SetStringValue("Publisher", "TaxiCheck")
	_ = k.SetStringValue("InstallLocation", dir)
	_ = k.SetStringValue("DisplayIcon", filepath.Join(dir, wininstall.BinName))
	_ = k.SetStringValue("UninstallString", `"`+uninstallPath+`"`)
	_ = k.SetDWordValue("NoModify", 1)
	_ = k.SetDWordValue("NoRepair", 1)
	if sizeKB > 0 && sizeKB < 1<<32 {
		_ = k.SetDWordValue("EstimatedSize", uint32(sizeKB))
	}
	return nil
}

func deleteUninstallReg() error {
	return registry.DeleteKey(registry.CURRENT_USER, uninstallRegKey)
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

func pause(in *bufio.Reader, silent bool) {
	if silent {
		return
	}
	fmt.Print("\nPress Enter to close...")
	_, _ = in.ReadString('\n')
}
