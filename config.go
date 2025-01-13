package main

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"
)

// Config application
type Config struct {
	Telegram struct {
		Bot         string `yaml:"bot" envconfig:"BOT_TOKEN"`
		SendRPSLimi int    `yaml:"send_rps_limit" default:"10" envconfig:"BOT_RATE_LIMIT"`
	} `yaml:"telegram"`

	App struct {
		ProfileDir string            `yaml:"profile_dir" default:"profiles" envconfig:"BOT_PROFILE_DIR"`
		FontsDir   string            `yaml:"fonts_dir" default:"fonts" envconfig:"BOT_FONTS_DIR"`
		Admins     []int64           `yaml:"admins,omitempty" envconfig:"BOT_ADMINS"`
		Fonts      map[string]string `yaml:"fonts,omitempty" envconfig:"BOT_FONT_DIR"`
	} `yaml:"app"`

	AI struct {
		ThreadsPerClient int `yaml:"threads_per_client" default:"6" envconfig:"BOT_THREADS_PER_CLIENT"`
		ThreadsPerAdmin  int `yaml:"threads_per_admin" default:"25" envconfig:"BOT_THREADS_PER_ADMIN"`
		WaitTimeout      int `yaml:"wait_timeout" default:"180" envconfig:"BOT_AI_TIMEOUT"`
	} `yaml:"ai"`
}

func loadConfig(name ...string) *Config {
	cfg := new(Config)
	defaults.SetDefaults(cfg)

	for _, n := range name {
		if cfgData, err := os.ReadFile(n); err != nil {
			panic(fmt.Sprintf("Can't load file %s: %s", n, err))
		} else {
			if err := yaml.Unmarshal(cfgData, cfg); err != nil {
				panic(fmt.Sprintf("Can't parse file %s: %s", n, err))
			}
		}
	}
	envconfig.Process("", cfg)
	return cfg
}

func (c *Config) String() string {
	res, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(res)
}
