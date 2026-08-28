package tui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	sourceRepoFile   = "source-repo"
	installedBinName = "taxiprijs"
)

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

	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		if dir := strings.TrimSpace(string(out)); isTaxiprijsRepo(dir) {
			return dir
		}
	}
	return ""
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

func applyNewBinaryImpl(dir string) (builtPath string, rebuildErr, installErr error) {
	built, err := rebuildBinary(dir)
	if err != nil {
		return "", err, nil
	}
	return built, nil, installBinary(built, dir)
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
