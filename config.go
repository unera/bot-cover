package main

import (
	"os"

	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"
)

// Config application
type Config struct {
	Telegram struct {
		Bot string `yaml:"bot"`
	} `yaml:"telegram"`

	App struct {
		ProfileDir string            `yaml:"profile_dir" default:"profiles"`
		FontsDir   string            `yaml:"fonts_dir" default:"fonts"`
		Admins     []int64           `yaml:"admins,omitempty"`
		Fonts      map[string]string `yaml:"fonts,omitempty"`
	} `yaml:"app"`

	AI struct {
		ThreadsPerClient int `yaml:"threads_per_client" default:"5"`
		ThreadsPerAdmin  int `yaml:"threads_per_admin" default:"25"`
	} `yaml:"ai"`
}

func loadConfig(name string) *Config {
	cfg := new(Config)
	defaults.SetDefaults(cfg)

	if cfgData, err := os.ReadFile(name); err != nil {
		panic(err)
	} else {
		if err := yaml.Unmarshal(cfgData, cfg); err != nil {
			panic(err)
		}
	}
	return cfg
}

func (c *Config) String() string {
	res, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(res)
}
