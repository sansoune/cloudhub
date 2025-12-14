package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	StackPath string `yaml:"stack_path"`

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

	cfg := DefaultConfig()
	if len(data) == 0 {
		return cfg, nil
	}


	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.ApplyDefaults()

	return cfg, nil
}

func (cfg *Config) ApplyDefaults() {
	if cfg.StackPath == "" {
		cfg.StackPath = "/opt/stack"
	}

	if cfg.Daemon.Interval == 0 {
		cfg.Daemon.Interval = 60
	}
}



func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.StackPath = "/opt/stack"
	cfg.Daemon.Interval = 60
	cfg.Notifications.Ntfy.Enabled = false
	cfg.Notifications.Ntfy.Server = ""
	cfg.Notifications.Ntfy.Topic = ""
	return cfg
}

func GetConfigPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "error working dir", err
	}

	cfgPath := filepath.Join(dir, "config.yaml")

	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath, nil
	}

	return "", nil
}
			
