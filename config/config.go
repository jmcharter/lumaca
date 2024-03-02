package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Directories struct {
		Posts     string
		Pages     string
		Static    string
		Templates string
		Dist      string
	}
	Author struct {
		Name string
	}
	Files struct {
		Extension string
	}
	Site struct {
		Title  string
		Author string
	}
}

func InitConfig() (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config file: %w", err)
	}
	return cfg, nil
}
