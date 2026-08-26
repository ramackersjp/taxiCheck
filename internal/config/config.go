package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
)

const (
	configDirName  = ".taxiprijs"
	configFileName = "config.toml"
)

type Config struct {
	Language        string           `toml:"language"`
	PassengerGroups []PassengerGroup `toml:"passenger_groups"`
}

type PassengerGroup struct {
	Name       string  `toml:"name"`
	BoardFee   float64 `toml:"board_fee"`
	PerKm      float64 `toml:"per_km"`
	PerMinute  float64 `toml:"per_minute"`
	WaitMinute float64 `toml:"wait_minute"`
}

func DefaultConfig() *Config {
	return &Config{
		PassengerGroups: []PassengerGroup{
			{
				Name:       "Taxi auto (max 4)",
				BoardFee:   4.31,
				PerKm:      3.17,
				PerMinute:  0.52,
				WaitMinute: 59.41,
			},
			{
				Name:       "Taxi bus (5-8)",
				BoardFee:   8.77,
				PerKm:      4.00,
				PerMinute:  0.59,
				WaitMinute: 59.41,
			},
		},
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get home directory: %w", err)
	}

	dir := filepath.Join(home, configDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory: %w", err)
	}

	return filepath.Join(dir, configFileName), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	cfg := &Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := validate(cfg); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}

	return nil
}

func validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	if len(cfg.PassengerGroups) == 0 {
		return fmt.Errorf("no passenger groups configured")
	}

	for i, g := range cfg.PassengerGroups {
		if g.Name == "" {
			return fmt.Errorf("passenger group %d has empty name", i+1)
		}
		if g.BoardFee < 0 {
			return fmt.Errorf("passenger group %q: board fee cannot be negative", g.Name)
		}
		if g.PerKm < 0 {
			return fmt.Errorf("passenger group %q: per km rate cannot be negative", g.Name)
		}
		if g.PerMinute < 0 {
			return fmt.Errorf("passenger group %q: per minute rate cannot be negative", g.Name)
		}
		if g.WaitMinute < 0 {
			return fmt.Errorf("passenger group %q: wait minute rate cannot be negative", g.Name)
		}
	}

	return nil
}

func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("226")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)
