package config

import (
	"errors"

	"github.com/BurntSushi/toml"
)

var DecodeFileError = errors.New("failed to decode config file")

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
		return Config{}, DecodeFileError
	}
	return cfg, nil
}
