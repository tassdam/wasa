package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ardanlabs/conf"
	"gopkg.in/yaml.v2"
)

type WebAPIConfiguration struct {
	Config struct {
		Path string `conf:"default:/conf/config.yml"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Debug bool
	DB    struct {
		Filename string `conf:"default:/tmp/decaf.db"`
	}
}

func loadConfiguration() (WebAPIConfiguration, error) {
	var cfg WebAPIConfiguration
	if err := conf.Parse(os.Args[1:], "CFG", &cfg); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			usage, err := conf.Usage("CFG", &cfg)
			if err != nil {
				return cfg, fmt.Errorf("generating config usage: %w", err)
			}
			fmt.Println(usage)
			return cfg, conf.ErrHelpWanted
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}
	fp, err := os.Open(cfg.Config.Path)
	if err != nil && !os.IsNotExist(err) {
		return cfg, fmt.Errorf("can't read the config file, while it exists: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return cfg, fmt.Errorf("can't read config file: %w", err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			return cfg, fmt.Errorf("can't unmarshal config file: %w", err)
		}
		_ = fp.Close()
	}
	return cfg, nil
}
