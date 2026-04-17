package config

import (
	"fmt"
	"os"
	"time"

	"github.com/wb-go/wbf/logger"
	"gopkg.in/yaml.v3"
)

// Config содержит общую конфигурацию приложения.
type Config struct {
	Server    ServerConfig    `yaml:"server" validate:"required"`
	Logger    LoggerConfig    `yaml:"logger" validate:"required"`
	Gin       GinConfig       `yaml:"gin" validate:"required"`
	Postgres  PostgresConfig  `yaml:"postgres" validate:"required"`
	Scheduler SchedulerConfig `yaml:"scheduler" validate:"required"`
}

// ServerConfig содержит настройки HTTP-сервера.
type ServerConfig struct {
	Addr         string        `yaml:"addr" env:"SERVER_ADDR" env-default:":8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

// LoggerConfig содержит настройки логирования.
type LoggerConfig struct {
	Engine string `yaml:"engine" env:"LOG_ENGINE" env-default:"slog"`
	Level  string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// LogLevel преобразует строковый уровень логирования в тип логгера.
func (c LoggerConfig) LogLevel() logger.Level {
	switch c.Level {
	case "debug":
		return logger.DebugLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return logger.InfoLevel
	}
}

// LogEngine преобразует строковое имя движка логирования в тип логгера.
func (c LoggerConfig) LogEngine() logger.Engine {
	return logger.Engine(c.Engine)
}

// GinConfig содержит настройки режима работы HTTP-роутера.
type GinConfig struct {
	Mode string `yaml:"mode" env:"GIN_MODE" env-default:"debug"`
}

// PostgresConfig содержит параметры подключения к PostgreSQL.
type PostgresConfig struct {
	Host            string        `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port            int           `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User            string        `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Database        string        `yaml:"database" env:"DB_NAME" env-default:"eventbooker"`
	SSLMode         string        `yaml:"sslmode" env:"DB_SSLMODE" env-default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" env-default:"5m"`
}

// DSN формирует строку подключения к PostgreSQL.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode,
	)
}

// SchedulerConfig содержит настройки фонового планировщика.
type SchedulerConfig struct {
	Interval time.Duration `yaml:"interval" env:"SCHEDULER_INTERVAL" env-default:"30s"`
}

// Load загружает конфигурацию приложения из YAML-файла по указанному пути.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
