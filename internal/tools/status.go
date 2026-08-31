// Package tools installs and reports Git, GitHub CLI, and GitHub login status.
package tools

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Overridable for tests.
var (
	lookPath = exec.LookPath
	runCmd   = func(name string, args ...string) ([]byte, error) {
		cmd := exec.Command(name, args...)
		cmd.Env = os.Environ()
		return cmd.CombinedOutput()
	}
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
)

// Status is the current Git / gh / GitHub login state.
type Status struct {
	Git      bool
	GH       bool
	LoggedIn bool
	User     string
}

// Probe returns whether git and gh are on PATH and whether gh is logged in.
func Probe() Status {
	EnsureToolPaths()
	s := Status{
		Git: GitInstalled(),
		GH:  GHInstalled(),
	}
	if s.GH {
		s.LoggedIn, s.User = GHUser()
	}
	return s
}

func GitInstalled() bool {
	if _, err := lookPath("git"); err == nil {
		return true
	}
	if goos == "windows" {
		if _, err := lookPath("git.exe"); err == nil {
			return true
		}
		for _, p := range windowsGitCandidates() {
			if fileExists(p) {
				return true
			}
		}
	}
	return false
}

func GHInstalled() bool {
	if _, err := lookPath("gh"); err == nil {
		return true
	}
	if goos == "windows" {
		if _, err := lookPath("gh.exe"); err == nil {
			return true
		}
		for _, p := range windowsGHCandidates() {
			if fileExists(p) {
				return true
			}
		}
	}
	return false
}

func GHUser() (bool, string) {
	out, err := runCmd("gh", "auth", "status")
	if err != nil {
		return false, ""
	}
	return parseGHAuth(string(out))
}

func parseGHAuth(out string) (bool, string) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "account "); i >= 0 {
			rest := strings.TrimSpace(line[i+len("account "):])
			user := strings.Fields(rest)
			if len(user) > 0 {
				return true, user[0]
			}
		}
	}
	if strings.Contains(out, "Logged in") {
		return true, ""
	}
	return false, ""
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func windowsGitCandidates() []string {
	var out []string
	for _, root := range windowsProgramRoots() {
		out = append(out, filepath.Join(root, "Git", "cmd", "git.exe"))
	}
	return out
}

func windowsGHCandidates() []string {
	var out []string
	for _, root := range windowsProgramRoots() {
		out = append(out, filepath.Join(root, "GitHub CLI", "gh.exe"))
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		out = append(out, filepath.Join(local, "GitHub CLI", "gh.exe"))
		out = append(out, filepath.Join(local, "Programs", "GitHub CLI", "gh.exe"))
	}
	return out
}

func windowsProgramRoots() []string {
	var roots []string
	for _, k := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if v := os.Getenv(k); v != "" {
			roots = append(roots, v)
		}
	}
	if len(roots) == 0 {
		roots = append(roots, `C:\Program Files`, `C:\Program Files (x86)`)
	}
	return roots
}

// EnsureToolPaths prepends common Git/gh install dirs so newly installed
// tools are found without restarting the app.
func EnsureToolPaths() {
	var extra []string
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		extra = append(extra, filepath.Join(home, ".local", "bin"))
	}
	if goos == "windows" {
		extra = append(extra, windowsGitDir(), windowsGHDir())
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			extra = append(extra, filepath.Join(local, "TaxiCheck"))
			extra = append(extra, filepath.Join(local, "GitHub CLI"))
		}
	}
	path := os.Getenv("PATH")
	for _, d := range extra {
		if d == "" || strings.Contains(path, d) {
			continue
		}
		path = d + string(os.PathListSeparator) + path
	}
	_ = os.Setenv("PATH", path)
}

func windowsGitDir() string {
	for _, p := range windowsGitCandidates() {
		if fileExists(p) {
			return filepath.Dir(p)
		}
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		return filepath.Join(pf, "Git", "cmd")
	}
	return `C:\Program Files\Git\cmd`
}

func windowsGHDir() string {
	for _, p := range windowsGHCandidates() {
		if fileExists(p) {
			return filepath.Dir(p)
		}
	}
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		return filepath.Join(pf, "GitHub CLI")
	}
	return `C:\Program Files\GitHub CLI`
}

func userLocalBin() (string, error) {
	if goos == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			dir := filepath.Join(local, "TaxiCheck")
			return dir, os.MkdirAll(dir, 0755)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "bin")
	return dir, os.MkdirAll(dir, 0755)
}
