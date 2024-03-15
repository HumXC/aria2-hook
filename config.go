package main

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Url                  string   `yaml:"url"`
	Token                string   `yaml:"token"`
	OnDownloadStart      []string `yaml:"onDownloadStart"`
	OnDownloadPause      []string `yaml:"onDownloadPause"`
	OnDownloadStop       []string `yaml:"onDownloadStop"`
	OnDownloadComplete   []string `yaml:"onDownloadComplete"`
	OnDownloadError      []string `yaml:"onDownloadError"`
	OnBtDownloadComplete []string `yaml:"onBtDownloadComplete"`
}

func DefaultConfig() Config {
	return Config{
		OnDownloadStart:      []string{},
		OnDownloadPause:      []string{},
		OnDownloadStop:       []string{},
		OnDownloadComplete:   []string{},
		OnDownloadError:      []string{},
		OnBtDownloadComplete: []string{},
	}
}

func ParseConfig(data []byte) (Config, error) {
	config := DefaultConfig()
	err := yaml.Unmarshal(data, &config)
	if strings.HasPrefix(config.Url, "http") {
		config.Url = strings.Replace(config.Url, "http", "ws", 1)
	}
	return config, err
}

func ParseConfigFile(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return DefaultConfig(), err
	}
	return ParseConfig(data)
}
