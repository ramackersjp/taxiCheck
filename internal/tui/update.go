package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	sourceRepoFile   = "source-repo"
	installedBinName = "taxiprijs"
	defaultRepoURL   = "https://github.com/ramackersjp/taxiCheck.git"
	sourceCloneDir   = "src"
)

// cloneSourceRepo downloads the TaxiCheck git repo when the app was installed
// without a checkout (Windows install.exe). Overridable in tests.
var cloneSourceRepo = cloneSourceRepoImpl

// Overridable for tests.
var (
	osExecutable = os.Executable
	osUserHome   = os.UserHomeDir
	osGetwd      = os.Getwd
	sudoInstall  = func(src, dst string) error {
		// -n never prompts: a TUI cannot show a sudo password prompt.
		return exec.Command("sudo", "-n", "install", "-Dm755", src, dst).Run()
	}
	systemInstallPath = "/usr/local/bin/" + installedBinName
)

func gitRepoDir() string {
	dir := locateGitRepo()
	if dir != "" {
		rememberRepoDir(dir)
		ensureRebuildHook(dir)
		return dir
	}
	return ""
}

func locateGitRepo() string {
	var candidates []string
	if wd, err := osGetwd(); err == nil && wd != "" {
		candidates = append(candidates, wd)
	}
	if exe, err := osExecutable(); err == nil && exe != "" {
		resolved := exe
		if r, err := filepath.EvalSymlinks(exe); err == nil && r != "" {
			resolved = r
		}
		candidates = append(candidates, filepath.Dir(resolved))
	}
	if saved := readRememberedRepoDir(); saved != "" {
		candidates = append(candidates, saved)
	}
	if home, err := osUserHome(); err == nil && home != "" {
		candidates = append(candidates, filepath.Join(home, ".taxiprijs", sourceCloneDir))
	}

	seen := map[string]bool{}
	for _, c := range candidates {
		c = filepath.Clean(c)
		if seen[c] {
			continue
		}
		seen[c] = true
		if dir := walkForGit(c); dir != "" && isTaxiprijsRepo(dir) {
			return dir
		}
	}

	if wd, err := osGetwd(); err == nil && wd != "" {
		if out, err := gitCmd(wd, "rev-parse", "--show-toplevel").Output(); err == nil {
			if dir := strings.TrimSpace(string(out)); isTaxiprijsRepo(dir) {
				return dir
			}
		}
	}
	return ""
}

func defaultSrcDir() string {
	home, err := osUserHome()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".taxiprijs", sourceCloneDir)
}

func gitBinary() string {
	if p, err := exec.LookPath("git"); err == nil {
		return p
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("git.exe"); err == nil {
			return p
		}
	}
	var candidates []string
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		candidates = append(candidates, filepath.Join(pf, "Git", "cmd", "git.exe"))
	}
	if pf := os.Getenv("ProgramFiles(x86)"); pf != "" {
		candidates = append(candidates, filepath.Join(pf, "Git", "cmd", "git.exe"))
	}
	if home, err := osUserHome(); err == nil && home != "" {
		candidates = append(candidates, filepath.Join(home, "AppData", "Local", "Programs", "Git", "cmd", "git.exe"))
	}
	candidates = append(candidates,
		`C:\Program Files\Git\cmd\git.exe`,
		"/usr/bin/git",
		"/usr/local/bin/git",
	)
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return "git"
}

func gitAvailable() bool {
	p := gitBinary()
	if filepath.Base(p) != p {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}
	_, err := exec.LookPath(p)
	return err == nil
}

// ensureSourceRepo returns a TaxiCheck git checkout, cloning it into
// ~/.taxiprijs/src when the Windows (or portable) install has no repo.
func ensureSourceRepo() (string, error) {
	if dir := gitRepoDir(); dir != "" {
		return dir, nil
	}
	return cloneSourceRepo()
}

func cloneSourceRepoImpl() (string, error) {
	if !gitAvailable() {
		return "", fmt.Errorf("git is not installed")
	}
	dest := defaultSrcDir()
	if dest == "" {
		return "", fmt.Errorf("cannot determine home directory")
	}
	if isTaxiprijsRepo(dest) {
		rememberRepoDir(dest)
		return dest, nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return "", err
	}
	if _, err := os.Stat(dest); err == nil {
		_ = os.RemoveAll(dest)
	}
	cmd := exec.Command(gitBinary(), "clone", defaultRepoURL, dest)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=true")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if !isTaxiprijsRepo(dest) {
		return "", fmt.Errorf("cloned repository is incomplete")
	}
	rememberRepoDir(dest)
	return dest, nil
}

func isTaxiprijsRepo(dir string) bool {
	if dir == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "taxiprijs")); err != nil {
		return false
	}
	return true
}

func walkForGit(start string) string {
	dir := start
	for {
		if dir == "" {
			return ""
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func rememberedRepoFile() string {
	home, err := osUserHome()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".taxiprijs", sourceRepoFile)
}

func readRememberedRepoDir() string {
	path := rememberedRepoFile()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	dir := strings.TrimSpace(string(data))
	if dir == "" {
		return ""
	}
	if !isTaxiprijsRepo(dir) {
		return ""
	}
	return dir
}

func rememberRepoDir(dir string) {
	path := rememberedRepoFile()
	if path == "" || dir == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	_ = os.WriteFile(path, []byte(dir+"\n"), 0644)
}

func goBinary() string {
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	for _, p := range []string{"/usr/bin/go", "/usr/local/go/bin/go", "/usr/lib/go/bin/go"} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return "go"
}

// applyNewBinary rebuilds from dir and installs the result. Overridable in tests.
var applyNewBinary = applyNewBinaryImpl

// runMakeInstall is overridable so tests do not invoke a real Makefile.
var runMakeInstall = runMakeInstallImpl

func applyNewBinaryImpl(dir string) (builtPath string, rebuildErr, installErr error) {
	built, err := rebuildBinary(dir)
	if err != nil {
		return "", err, nil
	}
	// Always place the binary on PATH first so a failed `make` still leaves
	// a runnable update.
	_ = installBinary(built, dir)
	if err := runMakeInstall(dir); err != nil {
		// `make install` copies the Omarchy QML plugin and (when sudo -n
		// works) the system binary. If make is missing, still copy QML.
		if qerr := copyOmarchyPlugin(dir, true); qerr != nil {
			return built, nil, err
		}
		return built, nil, nil
	}
	return built, nil, nil
}

// applyBranchBinary rebuilds and installs without `make install`, so switching
// branches never prompts for a sudo password (old release Makefiles run
// `sudo install` interactively). Overridable in tests.
var applyBranchBinary = applyBranchBinaryImpl

func applyBranchBinaryImpl(dir string) (builtPath string, rebuildErr, installErr error) {
	built, err := rebuildBinary(dir)
	if err != nil {
		return "", err, nil
	}
	_ = installBinary(built, dir)
	_ = copyOmarchyPlugin(dir, false)
	return built, nil, nil
}

const rebuildHookMarker = "taxiprijs-rebuild-hook"

// ensureRebuildHook installs a post-checkout hook so even an older TUI
// (v1.0.0) that only runs git checkout still rebuilds the user binary
// after switching back to dev.
func ensureRebuildHook(repo string) {
	if repo == "" {
		return
	}
	hookDir := filepath.Join(repo, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return
	}
	body := `#!/bin/sh
# ` + rebuildHookMarker + `
# Rebuild the user binary after a branch checkout so switching back to
# dev actually runs the dev menu (old app versions do not relaunch).
if [ "$3" != "1" ]; then
  exit 0
fi
repo=$(git rev-parse --show-toplevel) || exit 0
cd "$repo" || exit 0
if [ ! -d cmd/taxiprijs ]; then
  exit 0
fi
go=go
command -v go >/dev/null 2>&1 || go=/usr/bin/go
"$go" build -o .taxiprijs.new ./cmd/taxiprijs || exit 0
chmod +x .taxiprijs.new
mv -f .taxiprijs.new taxiprijs
if [ -n "$HOME" ]; then
  mkdir -p "$HOME/.local/bin"
  cp taxiprijs "$HOME/.local/bin/taxiprijs" 2>/dev/null || true
fi
`
	path := filepath.Join(hookDir, "post-checkout")
	if data, err := os.ReadFile(path); err == nil && !strings.Contains(string(data), rebuildHookMarker) {
		return
	}
	_ = os.WriteFile(path, []byte(body), 0755)
}

func runMakeInstallImpl(dir string) error {
	makeBin, err := exec.LookPath("make")
	if err != nil {
		return fmt.Errorf("make not found")
	}
	cmd := exec.Command(makeBin, "install")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func installOmarchyPlugin(dir string) error {
	return copyOmarchyPlugin(dir, true)
}

func copyOmarchyPlugin(dir string, restart bool) error {
	home, err := osUserHome()
	if err != nil || home == "" {
		return nil
	}
	src := filepath.Join(dir, "extras", "omarchy-plugin")
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	dst := filepath.Join(home, ".config", "omarchy", "plugins", "jp.taxiprijs")
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	for _, name := range []string{"BarWidget.qml", "manifest.json", "README.md"} {
		from := filepath.Join(src, name)
		if _, err := os.Stat(from); err != nil {
			continue
		}
		if err := copyRegular(from, filepath.Join(dst, name)); err != nil {
			return err
		}
	}
	if restart {
		restartOmarchyShell()
	}
	return nil
}

func restartOmarchyShell() {
	if _, err := exec.LookPath("omarchy"); err == nil {
		_ = exec.Command("omarchy", "restart", "shell").Run()
		return
	}
	if _, err := exec.LookPath("omarchy-shell"); err == nil {
		_ = exec.Command("omarchy-shell", "-q", "shell", "rescanPlugins").Run()
	}
}

func copyRegular(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".taxiprijs-copy-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	mode := info.Mode().Perm()
	if mode == 0 {
		mode = 0644
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dst)
}

// rebuildBinary compiles into a temp file, then atomically replaces the
// checkout binary. `go build -o taxiprijs` on a running binary fails with
// ETXTBSY ("text file busy") — that is how the Omarchy widget launches this
// app — so the temp+rename path is required for in-app updates.
func rebuildBinary(dir string) (string, error) {
	dest := filepath.Join(dir, installedBinName)
	tmp, err := os.CreateTemp(dir, ".taxiprijs-build-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	defer os.Remove(tmpName)

	build := exec.Command(goBinary(), "build", "-o", tmpName, "./cmd/taxiprijs")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	if err := os.Chmod(tmpName, 0755); err != nil {
		return "", err
	}
	if err := replaceFile(tmpName, dest); err != nil {
		return "", err
	}
	return dest, nil
}

func absClean(p string) string {
	if p == "" {
		return ""
	}
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return filepath.Clean(p)
}

func runningExecutable() string {
	exe, err := osExecutable()
	if err != nil || exe == "" {
		return ""
	}
	if r, err := filepath.EvalSymlinks(exe); err == nil && r != "" {
		return absClean(r)
	}
	return absClean(exe)
}

func installDest(src, dest string) error {
	if dest == "" {
		return fmt.Errorf("empty destination")
	}
	if err := replaceFile(src, dest); err == nil {
		return nil
	}
	return sudoInstall(src, dest)
}

// installBinary places the freshly built binary on every path the user might
// launch: the running executable, ~/.local/bin (PATH), and /usr/local/bin
// (make install / desktop entry) when writable or via passwordless sudo.
func installBinary(builtPath, repoDir string) error {
	if _, err := os.Stat(builtPath); err != nil {
		return err
	}
	builtAbs := absClean(builtPath)

	localBin := ""
	if home, err := osUserHome(); err == nil && home != "" {
		localBin = absClean(filepath.Join(home, ".local", "bin", installedBinName))
	}
	running := runningExecutable()
	systemBin := absClean(systemInstallPath)

	var firstErr error
	try := func(dest string) bool {
		if dest == "" || dest == builtAbs {
			return dest != ""
		}
		if err := installDest(builtPath, dest); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return false
		}
		return true
	}

	// ~/.local/bin is what `taxiprijs` on PATH resolves to for this user.
	localOK := localBin == "" || try(localBin)
	if running != "" && running != localBin {
		_ = try(running)
	}
	if systemBin != running && systemBin != localBin {
		_ = try(systemBin)
	}

	if !localOK {
		if firstErr != nil {
			return firstErr
		}
		return fmt.Errorf("could not install the new binary")
	}
	return nil
}

// copyFile copies src to dst and makes dst executable. The destination is
// replaced atomically (write to a temp file, then rename) so a running
// binary is not truncated.
func copyFile(src, dst string) error {
	return replaceFile(src, dst)
}

func replaceFile(src, dst string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	if srcAbs == dstAbs {
		return nil
	}

	in, err := os.Open(srcAbs)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dstAbs), 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(dstAbs), ".taxiprijs-new-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dstAbs)
}
