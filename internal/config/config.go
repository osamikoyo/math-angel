package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr     string        `yaml:"addr"`
	DBpath   string        `yaml:"db_path"`
	Timeout  time.Duration `yaml:"timeout"`
	Redis    Redis         `yaml:"redis"`
	Importer Importer      `yaml:"importer"`
}

type Redis struct {
	Addr    string        `yaml:"addr"`
	ExpTime time.Duration `yaml:"exp_time"`
}

type Importer struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Addr:    "localhost:8080",
		DBpath:  "storage/database.db",
		Timeout: 30 * time.Second,
		Redis: Redis{
			Addr:    "localhost:6379",
			ExpTime: 15 * time.Minute,
		},
		Importer: Importer{
			Enabled: false,
			File:    "math-source/test_source.jsonl",
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read config file, using defaults: %v", err)
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return cfg, nil
}
