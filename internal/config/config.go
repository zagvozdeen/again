package config

import (
	"encoding/json/v2"
	"fmt"
	"os"
)

type Config struct {
	App AppConfig      `json:"app"`
	DB  DatabaseConfig `json:"database"`
}

type AppConfig struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	IsProduction   bool   `json:"is_production"`
	DownMigrations bool   `json:"down_migrations"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func New() (*Config, error) {
	b, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg, json.RejectUnknownMembers(true))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config.json: %w", err)
	}
	return &cfg, nil
}
