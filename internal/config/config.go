// Package config manages itzd local configuration stored in ~/.itzd/config.json
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultAPIEndpoint         = "https://api.itz.agency"
	DefaultValidatorEndpoint   = "1.1.1.1:853"
	DefaultValidatorServerName = "cloudflare-dns.com"
	DefaultZone                = "itz.agency"
	configFileName             = "config.json"
)

// Config holds the local itzd configuration.
type Config struct {
	APIEndpoint         string `json:"api_endpoint"`
	Token               string `json:"token,omitempty"`
	WalletAddr          string `json:"wallet_address,omitempty"`
	Network             string `json:"network,omitempty"` // ethereum, solana, etc.
	ValidatorEndpoint   string `json:"validator_endpoint,omitempty"`
	ValidatorServerName string `json:"validator_server_name,omitempty"`
	DefaultZone         string `json:"default_zone,omitempty"`
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
		return withDefaults(&Config{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return withDefaults(&cfg), nil
}

func withDefaults(cfg *Config) *Config {
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = DefaultAPIEndpoint
	}
	if cfg.ValidatorEndpoint == "" {
		cfg.ValidatorEndpoint = DefaultValidatorEndpoint
	}
	if cfg.ValidatorServerName == "" {
		cfg.ValidatorServerName = DefaultValidatorServerName
	}
	if cfg.DefaultZone == "" {
		cfg.DefaultZone = DefaultZone
	}
	return cfg
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
