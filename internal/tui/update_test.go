package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMain(m *testing.M) {
	execReplacedProcess = func(string) error {
		return fmt.Errorf("relaunch disabled in tests")
	}
	os.Exit(m.Run())
}

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

func TestRebuildBinaryReplacesExistingOutput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tp\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmdDir := filepath.Join(dir, "cmd", "taxiprijs")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "taxiprijs")
	if err := os.WriteFile(dest, []byte("OLD-RUNNING-BINARY"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := rebuildBinary(dir)
	if err != nil {
		t.Fatalf("rebuildBinary: %v", err)
	}
	if got != dest {
		t.Fatalf("path = %q, want %q", got, dest)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "OLD-RUNNING-BINARY" {
		t.Fatal("rebuild left the previous binary in place")
	}
	if info, err := os.Stat(dest); err != nil || info.Mode().Perm()&0111 == 0 {
		t.Fatal("rebuilt binary must be executable")
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

func TestUpdateSuccessRelaunchesNewBinary(t *testing.T) {
	var got string
	orig := execReplacedProcess
	origExe := osExecutable
	execReplacedProcess = func(bin string) error {
		got = bin
		return nil
	}
	osExecutable = func() (string, error) { return "", fmt.Errorf("none") }
	defer func() {
		execReplacedProcess = orig
		osExecutable = origExe
	}()

	m := Model{screen: screenUpdate, lang: "en"}
	updated, cmd := m.Update(updateResultMsg{success: true, builtPath: "/tmp/taxiprijs"})
	if got != "/tmp/taxiprijs" {
		t.Fatalf("relaunch path = %q", got)
	}
	if cmd == nil {
		t.Fatal("expected Quit after a successful relaunch")
	}
	_ = updated
}

func TestF3ReinstallsWithoutPull(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	var called string
	orig := applyNewBinary
	applyNewBinary = func(d string) (string, error, error) {
		called = d
		return filepath.Join(d, "taxiprijs"), nil, nil
	}
	defer func() { applyNewBinary = orig }()

	m := Model{screen: screenUpdate, lang: "en", width: 80, height: 40}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyF3})
	if cmd == nil {
		t.Fatal("F3 must start a reinstall")
	}
	msg := cmd()
	res, ok := msg.(asyncMsg)
	if !ok {
		t.Fatalf("got %T, want asyncMsg", msg)
	}
	inner, ok := res.inner.(updateResultMsg)
	if !ok {
		t.Fatalf("inner %T, want updateResultMsg", res.inner)
	}
	if !inner.reinstall {
		t.Fatal("F3 must mark the result as a reinstall, not a git pull")
	}
	if called != dir {
		t.Fatalf("applyNewBinary dir=%q want %q", called, dir)
	}
	if inner.builtPath == "" {
		t.Fatal("expected builtPath")
	}
	_ = updated
}

func TestUpdateScreenShowsF3(t *testing.T) {
	m := Model{screen: screenUpdate, lang: "en", width: 80, height: 40, updateChecked: true}
	out := m.View()
	if !strings.Contains(out, "F3") {
		t.Fatalf("update screen must advertise F3, got:\n%s", out)
	}
}

func TestInstallOmarchyPluginCopiesQML(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "extras", "omarchy-plugin")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "BarWidget.qml"), []byte("/* qml */"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "manifest.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	origHome := osUserHome
	osUserHome = func() (string, error) { return home, nil }
	defer func() { osUserHome = origHome }()

	if err := installOmarchyPlugin(dir); err != nil {
		t.Fatalf("installOmarchyPlugin: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(home, ".config", "omarchy", "plugins", "jp.taxiprijs", "BarWidget.qml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "/* qml */" {
		t.Fatalf("QML content = %q", got)
	}
}

func TestApplyNewBinaryRunsMakeInstall(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tp\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmdDir := filepath.Join(dir, "cmd", "taxiprijs")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var makeDir string
	origMake := runMakeInstall
	origHome := osUserHome
	origExe := osExecutable
	home := t.TempDir()
	runMakeInstall = func(d string) error {
		makeDir = d
		return nil
	}
	osUserHome = func() (string, error) { return home, nil }
	osExecutable = func() (string, error) { return filepath.Join(home, "taxiprijs"), nil }
	defer func() {
		runMakeInstall = origMake
		osUserHome = origHome
		osExecutable = origExe
	}()

	built, rebuildErr, installErr := applyNewBinaryImpl(dir)
	if rebuildErr != nil {
		t.Fatalf("rebuild: %v", rebuildErr)
	}
	if installErr != nil {
		t.Fatalf("install: %v", installErr)
	}
	if makeDir != dir {
		t.Fatalf("make install dir=%q want %q", makeDir, dir)
	}
	if _, err := os.Stat(built); err != nil {
		t.Fatalf("built binary missing: %v", err)
	}
}

func TestPullUpdateAppliesNewBinary(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	var called string
	orig := applyNewBinary
	applyNewBinary = func(d string) (string, error, error) {
		called = d
		return filepath.Join(d, "taxiprijs"), nil, nil
	}
	defer func() { applyNewBinary = orig }()

	msg := Model{lang: "en"}.pullUpdate()()
	res, ok := msg.(updateResultMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if res.err != nil {
		t.Fatalf("pull: %v", res.err)
	}
	if called != dir {
		t.Fatalf("applyNewBinary dir=%q want %q", called, dir)
	}
	if res.builtPath == "" {
		t.Fatal("expected builtPath so the UI can relaunch")
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
