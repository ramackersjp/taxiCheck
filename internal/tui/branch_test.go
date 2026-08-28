package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRefNamesStripsHeadsPrefix(t *testing.T) {
	got := parseRefNames("refs/heads/dev\nrefs/heads/v1.0.1\nrefs/heads/v1.0.0\n", "refs/heads/")
	want := []string{"dev", "v1.0.1", "v1.0.0"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestParseRefNamesSkipsHEADAndEmpty(t *testing.T) {
	got := parseRefNames("\nrefs/remotes/origin/HEAD\nrefs/remotes/origin/v1.0.1\n", "refs/remotes/origin/")
	if len(got) != 1 || got[0] != "v1.0.1" {
		t.Fatalf("got %v, want [v1.0.1]", got)
	}
}

func TestParseRefNamesDoesNotKeepAmbiguousShortName(t *testing.T) {
	// %(refname:short) for a branch that shares a name with a tag is
	// "heads/v1.0.1", which must not be treated as a stable branch.
	if stableBranch("heads/v1.0.1") {
		t.Fatal("heads/v1.0.1 must not match stableBranch")
	}
	got := parseRefNames("refs/heads/v1.0.1\n", "refs/heads/")
	if len(got) != 1 || got[0] != "v1.0.1" {
		t.Fatalf("got %v, want [v1.0.1]", got)
	}
	if !stableBranch(got[0]) {
		t.Fatal("stripped name must match stableBranch")
	}
}

func TestSortSwitchableBranchesDevThenNewest(t *testing.T) {
	got := []string{"v1.0.0", "dev", "v1.0.1", "v2.0.0"}
	sortSwitchableBranches(got)
	want := []string{"dev", "v2.0.0", "v1.0.1", "v1.0.0"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCurrentGitBranchUnambiguousWithTag(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	gitIn(t, dir, "checkout", "v1.0.1")

	got := currentGitBranch(dir)
	if got != "v1.0.1" {
		t.Fatalf("currentGitBranch = %q, want v1.0.1", got)
	}
}

func TestDetectVersionWithTagAndBranch(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	gitIn(t, dir, "checkout", "v1.0.1")
	restoreRepoOverrides := overrideRepoLookup(t, dir)
	defer restoreRepoOverrides()

	orig := appVersion
	defer func() { appVersion = orig }()
	appVersion = "dev"
	DetectVersion()
	if appVersion != "v1.0.1" {
		t.Fatalf("appVersion = %q, want v1.0.1", appVersion)
	}
}

func TestFetchBranchesListsVersionNotHeadsPrefix(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	msg := Model{}.fetchBranches()()
	res, ok := msg.(branchResultMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if res.err != nil {
		t.Fatalf("fetch: %v", res.err)
	}
	if res.current != "dev" {
		t.Fatalf("current = %q, want dev", res.current)
	}
	found := false
	for _, b := range res.branches {
		if b == "heads/v1.0.1" {
			t.Fatal("listed ambiguous short name heads/v1.0.1")
		}
		if b == "v1.0.1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("v1.0.1 missing from %v", res.branches)
	}
	if len(res.branches) < 2 || res.branches[0] != "dev" {
		t.Fatalf("dev should be first, got %v", res.branches)
	}
}

func TestSwitchBranchRebuildsAndInstalls(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	var called string
	orig := applyBranchBinary
	applyBranchBinary = func(d string) (string, error, error) {
		called = d
		return filepath.Join(d, "taxiprijs"), nil, nil
	}
	defer func() { applyBranchBinary = orig }()

	msg := Model{}.switchBranch("v1.0.1")()
	res, ok := msg.(branchSwitchMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if res.err != nil {
		t.Fatalf("switch: %v", res.err)
	}
	if called != dir {
		t.Fatalf("applyBranchBinary dir=%q want %q (must rebuild this branch)", called, dir)
	}
	if res.builtPath == "" {
		t.Fatal("expected builtPath so the UI can relaunch")
	}
	got := strings.TrimSpace(gitIn(t, dir, "branch", "--show-current"))
	if got != "v1.0.1" {
		t.Fatalf("HEAD = %q, want v1.0.1", got)
	}
}

func TestSwitchBranchDoesNotRunMakeInstall(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	origApply := applyNewBinary
	origMake := runMakeInstall
	applyNewBinary = func(string) (string, error, error) {
		t.Fatal("branch switch must not call applyNewBinary (make install / sudo)")
		return "", nil, nil
	}
	runMakeInstall = func(string) error {
		t.Fatal("branch switch must not run make install")
		return nil
	}
	defer func() {
		applyNewBinary = origApply
		runMakeInstall = origMake
	}()

	msg := Model{}.switchBranch("v1.0.1")()
	res := msg.(branchSwitchMsg)
	if res.err != nil {
		t.Fatalf("switch: %v", res.err)
	}
}

func TestHelpExplainsPullThenF3(t *testing.T) {
	m := sizedModel("nl")
	m.screen = screenHelp
	out := m.View()
	if !strings.Contains(out, "eerst pullen") && !strings.Contains(out, "Eerst pull") {
		if !strings.Contains(out, "druk U om te pullen") {
			t.Fatalf("help must explain pull then F3, got:\n%s", out)
		}
	}
	if !strings.Contains(out, "F3") {
		t.Fatalf("help must mention F3, got:\n%s", out)
	}
}

func TestSwitchRelaunchesNewBinary(t *testing.T) {
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

	m := Model{screen: screenBranch, lang: "en"}
	updated, cmd := m.Update(branchSwitchMsg{success: true, newTag: "v1.0.0", builtPath: "/tmp/taxiprijs"})
	mm := updated.(Model)
	if mm.branchStatus == "" {
		t.Fatal("expected a success status")
	}
	if got != "/tmp/taxiprijs" {
		t.Fatalf("relaunch path = %q", got)
	}
	if cmd == nil {
		t.Fatal("expected Quit after a successful relaunch")
	}
}

func TestSwitchBranchWhenTagAndBranchShareName(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	msg := Model{}.switchBranch("v1.0.1")()
	res, ok := msg.(branchSwitchMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if res.err != nil {
		t.Fatalf("switch: %v", res.err)
	}
	if res.newTag != "v1.0.1" {
		t.Fatalf("newTag = %q", res.newTag)
	}
	got := strings.TrimSpace(gitIn(t, dir, "branch", "--show-current"))
	if got != "v1.0.1" {
		t.Fatalf("HEAD = %q, want v1.0.1", got)
	}
}

func TestSwitchBranchFallsBackWhenLocalAlreadyExists(t *testing.T) {
	dir := initAmbiguousVersionRepo(t)
	restore := overrideRepoLookup(t, dir)
	defer restore()

	if !gitRefExists(dir, "refs/heads/v1.0.1") {
		t.Fatal("expected local v1.0.1 to exist")
	}
	// Same as the in-app path: creating -b would fail with
	// "a branch named 'v1.0.1' already exists".
	msg := Model{}.switchBranch("v1.0.1")()
	res := msg.(branchSwitchMsg)
	if res.err != nil {
		t.Fatalf("expected checkout of existing branch, got %v", res.err)
	}
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// initAmbiguousVersionRepo is a taxiprijs-shaped repo on `dev` with both a
// branch and a tag named v1.0.1 — the situation that made in-app switch fail
// with "a branch named 'v1.0.1' already exists".
func initAmbiguousVersionRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-b", "dev")
	gitIn(t, dir, "config", "user.email", "test@example.com")
	gitIn(t, dir, "config", "user.name", "test")
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "taxiprijs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "taxiprijs", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", ".")
	gitIn(t, dir, "commit", "-m", "init")
	gitIn(t, dir, "branch", "v1.0.1")
	gitIn(t, dir, "tag", "v1.0.1")

	remote := t.TempDir()
	gitIn(t, remote, "init", "--bare")
	gitIn(t, dir, "remote", "add", "origin", remote)
	gitIn(t, dir, "push", "-u", "origin", "refs/heads/dev:refs/heads/dev")
	gitIn(t, dir, "push", "origin", "refs/heads/v1.0.1:refs/heads/v1.0.1")
	gitIn(t, dir, "fetch", "origin")
	return dir
}

func overrideRepoLookup(t *testing.T, dir string) func() {
	t.Helper()
	origExe, origHome, origWd := osExecutable, osUserHome, osGetwd
	origApply := applyNewBinary
	origBranch := applyBranchBinary
	home := t.TempDir()
	elsewhere := t.TempDir()
	osGetwd = func() (string, error) { return dir, nil }
	osUserHome = func() (string, error) { return home, nil }
	osExecutable = func() (string, error) { return filepath.Join(elsewhere, "taxiprijs"), nil }
	stub := func(string) (string, error, error) {
		return filepath.Join(dir, "taxiprijs"), nil, nil
	}
	applyNewBinary = stub
	applyBranchBinary = stub
	return func() {
		osExecutable = origExe
		osUserHome = origHome
		osGetwd = origWd
		applyNewBinary = origApply
		applyBranchBinary = origBranch
	}
}
