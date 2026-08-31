package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InstallGit installs Git using the platform package manager, or a silent
// Windows installer as fallback.
func InstallGit() error {
	EnsureToolPaths()
	if GitInstalled() {
		return nil
	}
	switch goos {
	case "windows":
		if err := wingetInstall("Git.Git"); err == nil {
			EnsureToolPaths()
			if GitInstalled() {
				return nil
			}
		}
		return fmt.Errorf("install Git via winget (Git.Git) or https://git-scm.com/download/win")
	case "darwin":
		if err := brewInstall("git"); err == nil {
			return nil
		}
		return fmt.Errorf("install Git with: brew install git")
	default:
		if err := linuxInstall("git"); err == nil {
			EnsureToolPaths()
			return nil
		}
		return fmt.Errorf("install Git with your package manager (pacman/apt/dnf install git)")
	}
}

// InstallGH installs the GitHub CLI (gh).
func InstallGH() error {
	EnsureToolPaths()
	if GHInstalled() {
		return nil
	}
	switch goos {
	case "windows":
		if err := wingetInstall("GitHub.cli"); err == nil {
			EnsureToolPaths()
			if GHInstalled() {
				return nil
			}
		}
		if err := downloadGH(); err == nil {
			return nil
		}
		return fmt.Errorf("install GitHub CLI via winget (GitHub.cli) or https://cli.github.com")
	case "darwin":
		if err := brewInstall("gh"); err == nil {
			return nil
		}
		if err := downloadGH(); err == nil {
			return nil
		}
		return fmt.Errorf("install GitHub CLI with: brew install gh")
	default:
		_ = linuxInstall("github-cli")
		if GHInstalled() {
			return nil
		}
		_ = linuxInstall("gh")
		if GHInstalled() {
			return nil
		}
		if err := downloadGH(); err == nil {
			return nil
		}
		return fmt.Errorf("install GitHub CLI with: pacman -S github-cli  or  apt install gh")
	}
}

// InstallMissing installs git and/or gh when they are absent.
func InstallMissing(s Status) error {
	var parts []string
	if !s.Git {
		if err := InstallGit(); err != nil {
			parts = append(parts, "git: "+err.Error())
		}
	}
	if !s.GH {
		if err := InstallGH(); err != nil {
			parts = append(parts, "gh: "+err.Error())
		}
	}
	EnsureToolPaths()
	if len(parts) > 0 {
		return fmt.Errorf("%s", strings.Join(parts, "; "))
	}
	return nil
}

func wingetInstall(id string) error {
	if _, err := lookPath("winget"); err != nil {
		return fmt.Errorf("winget not found")
	}
	out, err := runCmd("winget", "install", "--id", id, "-e", "--source", "winget",
		"--accept-package-agreements", "--accept-source-agreements", "--disable-interactivity")
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func brewInstall(pkg string) error {
	if _, err := lookPath("brew"); err != nil {
		return fmt.Errorf("brew not found")
	}
	out, err := runCmd("brew", "install", pkg)
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func linuxInstall(pkg string) error {
	type mgr struct {
		bin  string
		args []string
	}
	managers := []mgr{
		{"pacman", []string{"-S", "--noconfirm", "--needed", pkg}},
		{"apt-get", []string{"install", "-y", pkg}},
		{"dnf", []string{"install", "-y", pkg}},
		{"zypper", []string{"--non-interactive", "install", pkg}},
	}
	for _, m := range managers {
		if _, err := lookPath(m.bin); err != nil {
			continue
		}
		args := append([]string{"-n", m.bin}, m.args...)
		if _, err := lookPath("sudo"); err == nil {
			out, err := runCmd("sudo", args...)
			if err == nil {
				return nil
			}
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		out, err := runCmd(m.bin, m.args...)
		if err == nil {
			return nil
		}
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return fmt.Errorf("no supported package manager found")
}

func linuxPkgManager() string {
	for _, b := range []string{"pacman", "apt-get", "dnf", "zypper"} {
		if _, err := lookPath(b); err == nil {
			return b
		}
	}
	return ""
}

func hasBin(name string) bool {
	_, err := lookPath(name)
	return err == nil
}

// OpenSignup opens the GitHub registration page in the default browser.
func OpenSignup() error {
	return OpenURL("https://github.com/signup")
}

// OpenURL opens url with the platform's default handler.
func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch goos {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// LoginCommand runs an interactive GitHub CLI browser login.
func LoginCommand() *exec.Cmd {
	cmd := exec.Command("gh", "auth", "login", "--hostname", "github.com", "--git-protocol", "https", "--web")
	cmd.Env = os.Environ()
	return cmd
}
