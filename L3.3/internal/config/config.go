package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config хранит конфигурацию приложения.
type Config struct {
	App      AppConfig      `yaml:"app"`
	Postgres PostgresConfig `yaml:"postgres"`
}

// AppConfig хранит настройки HTTP-сервера.
type AppConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// PostgresConfig хранит настройки PostgreSQL.
type PostgresConfig struct {
	DSN string `yaml:"dsn"`
}

// Load читает конфигурацию из yaml-файла и переопределяет ее через env.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	overrideFromEnv(&cfg)

	return &cfg, nil
}

func overrideFromEnv(cfg *Config) {
	if value := os.Getenv("APP_HOST"); value != "" {
		cfg.App.Host = value
	}

	if value := os.Getenv("APP_PORT"); value != "" {
		cfg.App.Port = value
	}

	if value := os.Getenv("POSTGRES_DSN"); value != "" {
		cfg.Postgres.DSN = value
	}
}
