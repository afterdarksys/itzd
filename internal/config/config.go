// Package config manages itzd local configuration stored in ~/.itzd/config.json
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultAPIEndpoint = "https://api.itz.agency"
	configFileName     = "config.json"
)

// Config holds the local itzd configuration.
type Config struct {
	APIEndpoint string `json:"api_endpoint"`
	Token       string `json:"token,omitempty"`
	WalletAddr  string `json:"wallet_address,omitempty"`
	Network     string `json:"network,omitempty"` // ethereum, solana, etc.
}

func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".itzd"), nil
}

func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, configFileName), nil
}

// Load reads the config from disk. Returns a default config if none exists.
func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{APIEndpoint: DefaultAPIEndpoint}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = DefaultAPIEndpoint
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0700); err != nil {
		return err
	}
	p := filepath.Join(d, configFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}
