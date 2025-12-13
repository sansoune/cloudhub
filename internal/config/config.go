package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Daemon struct {
		Interval int `yaml:"interval"`
	} `yaml:"daemon"`

	Notifications struct {
		Ntfy struct {
			Enabled bool `yaml:"enabled"`
			Server string `yaml:"server"`
			Topic string `yaml:"topic"`
		} `yaml:"ntfy"`
	} `yaml:"notifications"`
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return  nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func GetConfigPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "error working dir", err
	}

	cfgPath := filepath.Join(dir, "config.yaml")
	fmt.Println("Current Working Directory:", cfgPath)

	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath, nil
	}

	return "", nil
}
			
