package config

import (
	"fmt"
	"os"
// "path/filepath"

// "gopkg.in/yaml.v3"
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

	fmt.Println(configPath)

	return nil, nil
}

func GetConfigPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "error working dir", err
	}

	fmt.Println("Current Working Directory:", dir)

	return dir, nil
}
			
