package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "bin")
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("content = %q, want %q", data, "hello")
	}
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Fatalf("dst is not executable: %v", info.Mode())
	}
}

func TestReplaceFileDoesNotTruncateInPlace(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("new-binary"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("old-running-binary"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
}

func TestWalkForGit(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	got := walkForGit(nested)
	if got != root {
		t.Fatalf("walkForGit = %q, want %q", got, root)
	}
	if walkForGit(t.TempDir()) != "" {
		t.Fatal("expected no repo in an empty temp dir")
	}
}

func TestLocateGitRepoUsesCwdAndSavedPath(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "cmd", "taxiprijs"), 0755); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	elsewhere := t.TempDir()

	origExe, origHome, origWd := osExecutable, osUserHome, osGetwd
	defer func() {
		osExecutable = origExe
		osUserHome = origHome
		osGetwd = origWd
	}()

	osExecutable = func() (string, error) { return filepath.Join(elsewhere, "taxiprijs"), nil }
	osUserHome = func() (string, error) { return home, nil }
	osGetwd = func() (string, error) { return repo, nil }

	got := locateGitRepo()
	if got != repo {
		t.Fatalf("from cwd: got %q want %q", got, repo)
	}

	osGetwd = func() (string, error) { return elsewhere, nil }
	rememberRepoDir(repo)
	got = locateGitRepo()
	if got != repo {
		t.Fatalf("from saved path: got %q want %q", got, repo)
	}
}

func TestInstallBinaryCopiesToLocalBin(t *testing.T) {
	home := t.TempDir()
	builtDir := t.TempDir()
	built := filepath.Join(builtDir, "taxiprijs")
	if err := os.WriteFile(built, []byte("fresh"), 0755); err != nil {
		t.Fatal(err)
	}
	sys := filepath.Join(t.TempDir(), "system-taxiprijs")
	running := filepath.Join(home, ".local", "bin", "taxiprijs")

	origExe, origHome, origSudo, origSys := osExecutable, osUserHome, sudoInstall, systemInstallPath
	defer func() {
		osExecutable = origExe
		osUserHome = origHome
		sudoInstall = origSudo
		systemInstallPath = origSys
	}()
	osUserHome = func() (string, error) { return home, nil }
	osExecutable = func() (string, error) { return running, nil }
	systemInstallPath = sys
	sudoCalled := false
	sudoInstall = func(src, dst string) error {
		sudoCalled = true
		return fmt.Errorf("sudo disabled in test")
	}

	if err := installBinary(built, builtDir); err != nil {
		t.Fatalf("installBinary: %v", err)
	}
	got, err := os.ReadFile(running)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "fresh" {
		t.Fatalf("local bin content = %q", got)
	}
	if sudoCalled && fileExists(sys) {
		t.Fatal("sudo should not have installed the system binary in this test")
	}
}

func TestInstallErrorDoesNotShowPaths(t *testing.T) {
	m := Model{screen: screenUpdate, lang: "en"}
	updated, _ := m.Update(updateResultMsg{
		success:    true,
		installErr: fmt.Errorf("open /usr/local/bin/taxiprijs: permission denied"),
	})
	mm := updated.(Model)
	if strings.Contains(mm.updateStatus, "/usr/local") || strings.Contains(mm.updateStatus, "permission denied") {
		t.Fatalf("install error leaked details: %q", mm.updateStatus)
	}
	if !strings.Contains(mm.updateStatus, "could not be installed") {
		t.Fatalf("expected generic install failure, got %q", mm.updateStatus)
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
