package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

type Config struct {
	User     string       `json:"user"`
	Password string       `json:"password,omitempty"`
	Token    oauth2.Token `json:"token,omitempty"`
}

func GetDefault() *Config {
	return &Config{}
}

func GetPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, "mchat")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return GetDefault(), nil

		}
		return nil, err
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) SaveConfig() error {
	path, err := GetPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0600)
	return err
}

func (c *Config) IsGoogle() bool {
	return c.Token.AccessToken != ""
}
