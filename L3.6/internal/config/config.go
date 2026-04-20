package config

import (
	"log"

	cleanenvport "github.com/wb-go/wbf/config/cleanenv-port"
)

// Config хранит конфигурацию приложения.
type Config struct {
	HTTP HTTPConfig `yaml:"http"`
	DB   DBConfig   `yaml:"db"`
}

// HTTPConfig хранит настройки HTTP-сервера.
type HTTPConfig struct {
	Port string `yaml:"port" env:"HTTP_PORT" validate:"required"`
	Mode string `yaml:"mode" env:"HTTP_MODE"`
}

// DBConfig хранит настройки подключения к PostgreSQL.
type DBConfig struct {
	MasterDSN string   `yaml:"master_dsn" env:"DB_MASTER_DSN" validate:"required"`
	SlaveDSNs []string `yaml:"slave_dsns" env:"DB_SLAVE_DSNS"`
}

// MustLoad загружает конфигурацию или завершает приложение с ошибкой.
func MustLoad() *Config {
	var cfg Config

	if err := cleanenvport.Load(&cfg); err != nil {
		log.Fatalf("ошибка загрузки конфигурации: %v", err)
	}

	return &cfg
}
