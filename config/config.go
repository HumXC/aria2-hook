package config

import (
	"context"
	"errors"
	"flag"
	"os"
	"reflect"
	"strings"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Url                  string   `yaml:"url" env:"URL"`
	Token                string   `yaml:"token" env:"TOKEN"`
	Debug                bool     `yaml:"debug" env:"DEBUG"`
	OnDownloadStart      []string `yaml:"onDownloadStart" env:"ON_DOWNLOAD_START, delimiter=#"`
	OnDownloadPause      []string `yaml:"onDownloadPause" env:"ON_DOWNLOAD_PAUSE, delimiter=#"`
	OnDownloadStop       []string `yaml:"onDownloadStop" env:"ON_DOWNLOAD_STOP, delimiter=#"`
	OnDownloadComplete   []string `yaml:"onDownloadComplete" env:"ON_DOWNLOAD_COMPLETE, delimiter=#"`
	OnDownloadError      []string `yaml:"onDownloadError" env:"ON_DOWNLOAD_ERROR, delimiter=#"`
	OnBtDownloadComplete []string `yaml:"onBtDownloadComplete" env:"ON_BT_DOWNLOAD_COMPLETE, delimiter=#"`
}

func Verify(cfg Config) error {
	if cfg.Url == "" {
		return errors.New("url is must be set")
	}
	return nil
}

func Merge(cfg ...Config) Config {
	m := func(c1, c2 Config) Config {
		v1 := reflect.ValueOf(&c1).Elem()
		v2 := reflect.ValueOf(&c2).Elem()
		for i := 0; i < v2.NumField(); i++ {
			if v2.Field(i).Kind() == reflect.Slice {
				v1.Field(i).Set(reflect.AppendSlice(v1.Field(i), v2.Field(i)))
				continue
			}
			if v2.Field(i).Kind() == reflect.String {
				if v2.Field(i).String() == "" {
					continue
				}
			}
			v1.Field(i).Set(v2.Field(i))
		}
		return c1
	}
	var config Config
	for _, c := range cfg {
		config = m(config, c)
	}
	return config
}

func Load(data []byte) (Config, error) {
	config := Config{}
	err := yaml.Unmarshal(data, &config)

	return config, err
}

func FromFile(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	return SFromFile(data)
}

func SFromFile(data []byte) (Config, error) {
	c := Config{}
	err := yaml.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func FromEnv() (Config, error) {
	c := Config{}
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return c, err
	}
	return c, nil
}

func FromCmd(cmd *flag.FlagSet) (Config, error) {
	return SFromCmd(cmd, os.Args)
}

func SFromCmd(cmd *flag.FlagSet, args []string) (Config, error) {
	c := Config{}
	cmd.StringVar(&c.Url, "url", "", "aria2 jsonrpc url")
	cmd.StringVar(&c.Token, "token", "", "aria2 jsonrpc token")
	cmd.BoolVar(&c.Debug, "debug", false, "debug mode")
	OnDownloadStart := ""
	OnDownloadPause := ""
	OnDownloadStop := ""
	OnDownloadComplete := ""
	OnDownloadError := ""
	OnBtDownloadComplete := ""

	cmd.StringVar(&OnDownloadStart, "on-download-start", "", "event: onDownloadStart")
	cmd.StringVar(&OnDownloadPause, "on-download-pause", "", "event: onDownloadPause")
	cmd.StringVar(&OnDownloadStop, "on-download-stop", "", "event: onDownloadStop")
	cmd.StringVar(&OnDownloadComplete, "on-download-complete", "", "event: onDownloadComplete")
	cmd.StringVar(&OnDownloadError, "on-download-error", "", "event: onDownloadError")
	cmd.StringVar(&OnBtDownloadComplete, "on-bt-download-complete", "", "event: onBtDownloadComplete")

	if err := cmd.Parse(args[1:]); err != nil {
		return c, err
	}
	argToSlice := func(s string) (result []string) {
		if s == "" {
			return
		}
		result = strings.Split(s, "#")
		return
	}
	c.OnDownloadStart = argToSlice(OnDownloadStart)
	c.OnDownloadPause = argToSlice(OnDownloadPause)
	c.OnDownloadStop = argToSlice(OnDownloadStop)
	c.OnDownloadComplete = argToSlice(OnDownloadComplete)
	c.OnDownloadError = argToSlice(OnDownloadError)
	c.OnBtDownloadComplete = argToSlice(OnBtDownloadComplete)

	return c, nil
}
