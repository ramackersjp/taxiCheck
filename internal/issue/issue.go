package issue

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Result reports the outcome of submitting an issue.
type Result struct {
	IssueNumber   int
	IssueURL      string
	RemoteSet     bool // a git remote is configured (points at a GitHub repo)
	GHAvailable   bool // the `gh` CLI is installed
	LoggedLocally bool
	LocalLogPath  string
	SkipReason    string // non-empty when nothing was created
}

// LogDir returns the directory where local issue logs are stored.
func LogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".taxiprijs", "logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// systemInfo builds a short, structured description of the running system so
// issues can be debugged across distros and operating systems.
func systemInfo() string {
	var b strings.Builder
	fmt.Fprintf(&b, "OS: %s\n", runtime.GOOS)
	fmt.Fprintf(&b, "Arch: %s\n", runtime.GOARCH)
	// Try to report a Linux distro (e.g. Arch, Ubuntu) when present.
	if runtime.GOOS == "linux" && fileExists("/etc/os-release") {
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					fmt.Fprintf(&b, "Distro: %s\n", strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`))
					break
				}
			}
		}
	}
	if data, err := os.ReadFile("/proc/version"); err == nil && runtime.GOOS == "linux" {
		fmt.Fprintf(&b, "Kernel: %s\n", strings.TrimSpace(strings.Split(string(data), "\n")[0]))
	}
	fmt.Fprintf(&b, "Go: %s\n", runtime.Version())
	return b.String()
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// composeBody builds the Markdown body for the issue: the user's description
// and error output plus collected system information.
func composeBody(description, errorOutput string) string {
	var b strings.Builder
	if description != "" {
		b.WriteString("## Description\n\n")
		b.WriteString(description)
		b.WriteString("\n\n")
	}
	if errorOutput != "" {
		b.WriteString("## Error output\n\n```\n")
		b.WriteString(errorOutput)
		b.WriteString("\n```\n\n")
	}
	b.WriteString("## System\n\n```\n")
	b.WriteString(systemInfo())
	b.WriteString("```\n")
	return b.String()
}

// writeLocalLog always stores the report on disk so nothing is ever lost, even
// when GitHub / the gh CLI is unavailable.
func writeLocalLog(description, errorOutput string) (string, error) {
	dir, err := LogDir()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("issue-%s.md", time.Now().Format("2006-01-02T15-04-05"))
	path := filepath.Join(dir, name)
	content := composeBody(description, errorOutput)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// ghAvailable reports whether the GitHub CLI is installed.
func ghAvailable() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// repoDir returns the local git repository directory, if the app runs from one.
func repoDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	for {
		if fileExists(filepath.Join(dir, ".git")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// remoteURL returns the first (fetch) remote URL, or "" if none is set.
func remoteURL() string {
	dir := repoDir()
	if dir == "" {
		return ""
	}
	out, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Report logs the issue locally and, if the GitHub CLI is installed and the
// repo has a GitHub remote, creates a GitHub issue and returns its number.
// If `gh` or GitHub is unavailable the app keeps working — the report is only
// written to the local log.
func Report(description, errorOutput string) Result {
	res := Result{}

	// 1. Always persist locally first.
	path, err := writeLocalLog(description, errorOutput)
	if err == nil {
		res.LoggedLocally = true
		res.LocalLogPath = path
	} else {
		res.SkipReason = "Failed to write local log: " + err.Error()
		return res
	}

	// 2. Determine whether we can create a GitHub issue.
	remote := remoteURL()
	site := siteFromRemote(remote)
	res.RemoteSet = site != ""

	res.GHAvailable = ghAvailable()
	if !res.GHAvailable {
		res.SkipReason = "GitHub CLI (gh) not installed. Issue saved to local log only."
		return res
	}
	if site == "" {
		res.SkipReason = "No GitHub remote configured. Issue saved to local log only."
		return res
	}

	// 3. Create the GitHub issue and extract its number.
	body := composeBody(description, errorOutput)
	title := strings.TrimSpace(description)
	if title == "" {
		title = "Bug report"
	}
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	args := []string{"issue", "create", "--title", title, "--body", body, "--repo", site}
	cmd := exec.Command("gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.SkipReason = fmt.Sprintf("gh issue create failed: %s (%s)", err, strings.TrimSpace(string(out)))
		return res
	}

	url := strings.TrimSpace(string(out))
	res.IssueURL = url
	res.IssueNumber = numberFromURL(url)
	return res
}

// siteFromRemote converts a git remote URL (HTTPS or SSH) into an owner/repo
// slug that `gh` can target, or "" if it is not a GitHub remote.
func siteFromRemote(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	var slug string
	if strings.HasPrefix(u, "git@github.com:") {
		slug = strings.TrimPrefix(u, "git@github.com:")
	} else if strings.HasPrefix(u, "https://github.com/") {
		slug = strings.TrimPrefix(u, "https://github.com/")
	} else if strings.HasPrefix(u, "ssh://git@github.com/") {
		slug = strings.TrimPrefix(u, "ssh://git@github.com/")
	} else {
		return ""
	}
	slug = strings.TrimSuffix(slug, ".git")
	seg := strings.Split(slug, "/")
	if len(seg) < 2 {
		return ""
	}
	return seg[0] + "/" + seg[1]
}

// numberFromURL extracts the trailing issue number from a gh issue URL.
func numberFromURL(url string) int {
	i := strings.LastIndex(url, "/")
	if i < 0 {
		return 0
	}
	var n int
	fmt.Sscanf(url[i+1:], "%d", &n)
	return n
}
