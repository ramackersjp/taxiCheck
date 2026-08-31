package wininstall

import "testing"

func TestPathContains(t *testing.T) {
	list := `C:\Windows;C:\Windows\System32;C:\Users\jp\AppData\Local\TaxiCheck`
	if !PathContains(list, `C:\Users\jp\AppData\Local\TaxiCheck`) {
		t.Fatal("expected exact dir to be found")
	}
	if !PathContains(list, `c:\users\jp\appdata\local\taxicheck\`) {
		t.Fatal("expected case-insensitive match with trailing slash")
	}
	if PathContains(list, `C:\Users\jp\AppData\Local\Other`) {
		t.Fatal("did not expect unrelated dir")
	}
	if PathContains("", `C:\TaxiCheck`) {
		t.Fatal("empty PATH should not contain a dir")
	}
}

func TestAppendPath(t *testing.T) {
	got := AppendPath(`C:\Windows;`, `C:\TaxiCheck`)
	if got != `C:\Windows;C:\TaxiCheck` {
		t.Fatalf("got %q", got)
	}
	same := AppendPath(got, `c:\taxicheck\`)
	if same != got {
		t.Fatalf("duplicate append: %q", same)
	}
	if got := AppendPath("", `C:\TaxiCheck`); got != `C:\TaxiCheck` {
		t.Fatalf("empty start: %q", got)
	}
}

func TestUserInstallDirUsesLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", `C:\Users\jp\AppData\Local`)
	got := UserInstallDir()
	if got == "" || got == `C:\Users\jp\AppData\Local` {
		t.Fatalf("got %q", got)
	}
	if UserInstallBin() == "" {
		t.Fatal("empty install bin")
	}
}

func TestRemovePath(t *testing.T) {
	list := `C:\Windows;C:\TaxiCheck;C:\Windows\System32`
	got := RemovePath(list, `c:\taxicheck\`)
	if got != `C:\Windows;C:\Windows\System32` {
		t.Fatalf("got %q", got)
	}
	if got := RemovePath(`C:\TaxiCheck`, `C:\TaxiCheck`); got != "" {
		t.Fatalf("remove last: %q", got)
	}
}
