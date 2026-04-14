package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config хранит конфигурацию приложения.
type Config struct {
	App struct {
		Port string `yaml:"port"`
	} `yaml:"app"`

	Postgres struct {
		DSN string `yaml:"dsn"`
	} `yaml:"postgres"`

	Kafka struct {
		Brokers  []string `yaml:"brokers"`
		Topic    string   `yaml:"topic"`
		GroupID  string   `yaml:"group_id"`
		DLQTopic string   `yaml:"dlq_topic"`
	} `yaml:"kafka"`

	Storage struct {
		OriginalDir  string `yaml:"original_dir"`
		ProcessedDir string `yaml:"processed_dir"`
		ThumbsDir    string `yaml:"thumbs_dir"`
	} `yaml:"storage"`

	Image struct {
		ResizeWidth   uint   `yaml:"resize_width"`
		ThumbWidth    uint   `yaml:"thumb_width"`
		ThumbHeight   uint   `yaml:"thumb_height"`
		WatermarkText string `yaml:"watermark_text"`
	} `yaml:"image"`
}

// Load загружает конфигурацию из yaml-файла.
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
