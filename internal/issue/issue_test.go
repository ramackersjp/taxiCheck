package issue

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoDirUsesRememberedSource(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".taxiprijs"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".taxiprijs", "source-repo"), []byte(repo+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := findRepoDir(home, t.TempDir(), filepath.Join(t.TempDir(), "taxiprijs"))
	if got != repo {
		t.Fatalf("got %q, want %q", got, repo)
	}
}

func TestSiteFromRemote(t *testing.T) {
	cases := map[string]string{
		"git@github.com:ramackersjp/taxiCheck.git":     "ramackersjp/taxiCheck",
		"https://github.com/ramackersjp/taxiCheck.git": "ramackersjp/taxiCheck",
		"https://github.com/ramackersjp/taxiCheck":     "ramackersjp/taxiCheck",
		"git@github.com:foo/bar":                       "foo/bar",
		"git@gitlab.com:foo/bar":                       "",
		"":                                             "",
	}
	for in, want := range cases {
		if got := siteFromRemote(in); got != want {
			t.Errorf("siteFromRemote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNumberFromURL(t *testing.T) {
	if got := numberFromURL("https://github.com/ramackersjp/taxiCheck/issues/42"); got != 42 {
		t.Errorf("numberFromURL = %d, want 42", got)
	}
	if got := numberFromURL("https://example.com/issue/not-a-number"); got != 0 {
		t.Errorf("numberFromURL = %d, want 0", got)
	}
}
