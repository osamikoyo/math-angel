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
	Auth     Auth          `yaml:"auth"`
}

type Redis struct {
	Addr    string        `yaml:"addr"`
	ExpTime time.Duration `yaml:"exp_time"`
}

type Importer struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file"`
}

type Auth struct {
	JWTSecretKey string        `yaml:"jwt_secret_key"`
	TokenTTL     time.Duration `yaml:"token_ttl"`
}

func LoadConfig(path string) (*Config, error) {
	log.Println("loading config")

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
		Auth: Auth{
			JWTSecretKey: "secret_key",
			TokenTTL:     30 * time.Minute,
		},
	}

	if path == "" {
		log.Println("config path is empty, using default config")
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("config file not found: %s, using default", path)
		} else {
			log.Printf("failed to read config file: %v, using default", err)
		}
		return cfg, nil
	}

	var newCfg Config
	if err := yaml.Unmarshal(data, &newCfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	log.Println("config loaded from file successfully")
	return &newCfg, nil
}
