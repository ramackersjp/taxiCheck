package wininstall

import "testing"

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantSilent    bool
		wantUninstall bool
	}{
		{name: "none", args: nil},
		{name: "slash S", args: []string{"/S"}, wantSilent: true},
		{name: "slash s", args: []string{"/s"}, wantSilent: true},
		{name: "dash silent", args: []string{"--silent"}, wantSilent: true},
		{name: "slash uninstall", args: []string{"/uninstall"}, wantUninstall: true},
		{name: "dash u", args: []string{"-u"}, wantUninstall: true},
		{name: "both", args: []string{"/S", "/uninstall"}, wantSilent: true, wantUninstall: true},
		{name: "unknown ignored", args: []string{"/help", "foo"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			silent, uninstall := ParseArgs(tt.args)
			if silent != tt.wantSilent || uninstall != tt.wantUninstall {
				t.Fatalf("ParseArgs(%q)=%t,%t want %t,%t", tt.args, silent, uninstall, tt.wantSilent, tt.wantUninstall)
			}
		})
	}
}

func TestDisplayVersion(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "dev"},
		{"dev", "dev"},
		{"1.1.0", "v1.1.0"},
		{"v1.1.0", "v1.1.0"},
		{" 1.1.0 ", "v1.1.0"},
	}
	for _, tt := range tests {
		if got := DisplayVersion(tt.in); got != tt.want {
			t.Fatalf("DisplayVersion(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}
