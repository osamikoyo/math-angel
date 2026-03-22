package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Addr     string   `mapstructure:"addr"`
	DBpath   string   `mapstructure:"db_path"`
	Redis    Redis    `mapstructure:"redis"`
	Importer Importer `mapstrcucture:"import"`
}

type Redis struct {
	Addr string `mapstructure:"addr"`
}

type Importer struct {
	Enabled bool   `mapstructure:"enabled"`
	File    string `mapstructure:"file"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("app.addr", "localhost:8080")
	v.SetDefault("app.do_import", false)
	v.SetDefault("app.db_path", "storage/database.db")

	v.SetDefault("app.redis.addr", "localhost:7079")

	v.SetDefault("app.importer.enabled", false)
	v.SetDefault("app.importer.file", "math-source/main_train.jsonl")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("fatal error reading config file: %w", err)
		}

		log.Printf("failed found config file, use default values")
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed parse config to struct")
	}

	return &cfg, nil
}
