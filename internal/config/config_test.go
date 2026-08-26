package config

import (
	"os"
	"path/filepath"
	"testing"
)

func cleanup(t *testing.T) {
	t.Helper()
	path, _ := ConfigPath()
	os.Remove(path)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.PassengerGroups) != 2 {
		t.Fatalf("expected 2 passenger groups, got %d", len(cfg.PassengerGroups))
	}

	if cfg.PassengerGroups[0].Name != "Taxi auto (max 4)" {
		t.Errorf("expected first group name 'Taxi auto (max 4)', got %q", cfg.PassengerGroups[0].Name)
	}

	if cfg.PassengerGroups[1].Name != "Taxi bus (5-8)" {
		t.Errorf("expected second group name 'Taxi bus (5-8)', got %q", cfg.PassengerGroups[1].Name)
	}
}

func TestValidateNilConfig(t *testing.T) {
	err := validate(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestValidateEmptyGroups(t *testing.T) {
	cfg := &Config{PassengerGroups: []PassengerGroup{}}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for empty passenger groups")
	}
}

func TestValidateNegativeValues(t *testing.T) {
	cfg := &Config{
		PassengerGroups: []PassengerGroup{
			{Name: "test", BoardFee: -1, PerKm: 3.0, PerMinute: 0.5, WaitMinute: 0.5},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for negative board fee")
	}
}

func TestValidateEmptyName(t *testing.T) {
	cfg := &Config{
		PassengerGroups: []PassengerGroup{
			{Name: "", BoardFee: 3.5, PerKm: 3.0, PerMinute: 0.5, WaitMinute: 0.5},
		},
	}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Ext(path) != ".toml" {
		t.Errorf("expected .toml extension, got %s", filepath.Ext(path))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	cfg := DefaultConfig()

	if err := Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	defer cleanup(t)

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded == nil {
		t.Fatal("loaded config is nil")
	}

	if len(loaded.PassengerGroups) != len(cfg.PassengerGroups) {
		t.Errorf("expected %d passenger groups, got %d", len(cfg.PassengerGroups), len(loaded.PassengerGroups))
	}

	for i, g := range loaded.PassengerGroups {
		if g.Name != cfg.PassengerGroups[i].Name {
			t.Errorf("group %d: expected name %q, got %q", i, cfg.PassengerGroups[i].Name, g.Name)
		}
		if g.BoardFee != cfg.PassengerGroups[i].BoardFee {
			t.Errorf("group %d: expected board fee %.2f, got %.2f", i, cfg.PassengerGroups[i].BoardFee, g.BoardFee)
		}
	}
}

func TestLoadNonexistentConfig(t *testing.T) {
	path, _ := ConfigPath()
	os.Remove(path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg != nil {
		t.Error("expected nil config for nonexistent file")
	}
}

func TestExists(t *testing.T) {
	cfg := DefaultConfig()
	Save(cfg)
	defer cleanup(t)

	if !Exists() {
		t.Error("expected config to exist after save")
	}
}
