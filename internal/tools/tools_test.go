package tools

import (
	"fmt"
	"os"
	"testing"
)

func TestParseGHAuth(t *testing.T) {
	out := "github.com\n  ✓ Logged in to github.com account ramackersjp (keyring)\n  - Active account: true\n"
	ok, user := parseGHAuth(out)
	if !ok || user != "ramackersjp" {
		t.Fatalf("ok=%v user=%q", ok, user)
	}
	ok, _ = parseGHAuth("not logged in")
	if ok {
		t.Fatal("expected not logged in")
	}
}

func TestProbeUsesLookPath(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(name string) (string, error) {
		if name == "git" || name == "gh" {
			return "/usr/bin/" + name, nil
		}
		return "", fmt.Errorf("not found")
	}
	origRun := runCmd
	runCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("github.com\n  ✓ Logged in to github.com account tester (keyring)\n"), nil
	}
	defer func() { runCmd = origRun }()

	s := Probe()
	if !s.Git || !s.GH {
		t.Fatalf("status=%+v", s)
	}
	if !s.LoggedIn || s.User != "tester" {
		t.Fatalf("login=%v user=%q", s.LoggedIn, s.User)
	}
}

func TestInstallMissingSkipsPresent(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(name string) (string, error) {
		return "/bin/" + name, nil
	}
	if err := InstallMissing(Status{Git: true, GH: true}); err != nil {
		t.Fatal(err)
	}
}

func TestLinuxPkgManager(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(name string) (string, error) {
		if name == "pacman" {
			return "/usr/bin/pacman", nil
		}
		return "", fmt.Errorf("missing")
	}
	if got := linuxPkgManager(); got != "pacman" {
		t.Fatalf("got %q", got)
	}
}

func TestHasBin(t *testing.T) {
	orig := lookPath
	defer func() { lookPath = orig }()
	lookPath = func(name string) (string, error) {
		if name == "git" {
			return "/usr/bin/git", nil
		}
		return "", os.ErrNotExist
	}
	if !hasBin("git") || hasBin("gh") {
		t.Fatal("hasBin mismatch")
	}
}
