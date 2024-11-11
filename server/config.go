package main

import (
	"fmt"
	"os"

	"github.com/sysarmor/guard/server/internal/repo/postgres"
	"github.com/sysarmor/guard/server/internal/service"
	"gopkg.in/yaml.v3"
)

type config struct {
	Addr string `yaml:"addr"`

	Postgres postgres.Config `yaml:"postgres"`

	Services service.Config `yaml:"services"`
}

func (c *config) Validate() error {
	if c.Addr == "" {
		c.Addr = ":80"
	}

	if err := c.Postgres.Validate(); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	if err := c.Services.Validate(); err != nil {
		return fmt.Errorf("services: %w", err)
	}

	return nil
}

func loadConfig(path string) (*config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer fd.Close()

	var cfg config
	if err := yaml.NewDecoder(fd).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}
