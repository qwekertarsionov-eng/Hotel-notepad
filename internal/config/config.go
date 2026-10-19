package config

import "os"

// Config holds application configuration loaded from environment variables.
type Config struct {
	DATABASE_URL string
	JWT_SECRET   string
	HTTP_ADDR    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		DATABASE_URL: os.Getenv("DATABASE_URL"),
		JWT_SECRET:   os.Getenv("JWT_SECRET"),
		HTTP_ADDR:    os.Getenv("HTTP_ADDR"),
	}
	if cfg.HTTP_ADDR == "" {
		cfg.HTTP_ADDR = ":8080"
	}
	return cfg, nil
}
