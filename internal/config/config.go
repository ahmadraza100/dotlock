package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user-level configuration for dotlock.
type Config struct {
	VaultFilename  string `json:"vault_filename"`
	KeysDir        string `json:"keys_dir"`
	DefaultProfile string `json:"default_profile"`
}

// DefaultConfig returns hardcoded defaults without reading disk.
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".config", "dotlock")
	return &Config{
		VaultFilename:  ".dotlock",
		KeysDir:        filepath.Join(cfgDir, "keys"),
		DefaultProfile: "dev",
	}
}

// LoadConfig reads persisted config from disk, falling back to defaults.
func LoadConfig() *Config {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "dotlock", "config.json")
	f, err := os.Open(path)
	if err != nil {
		return DefaultConfig()
	}
	defer f.Close()
	cfg := DefaultConfig()
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return DefaultConfig()
	}
	return cfg
}

// SaveConfig writes config to ~/.config/dotlock/config.json.
func SaveConfig(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfgDir := filepath.Join(home, ".config", "dotlock")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(cfgDir, "config.json")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}
