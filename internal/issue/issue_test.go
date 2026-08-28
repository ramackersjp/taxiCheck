package issue

import "testing"

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
