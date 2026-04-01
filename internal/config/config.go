package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Providers ProviderConfig
	Health    HealthCheckConfig
	Logging   LoggingConfig
}

type ServerConfig struct {
	Port         string        `env:"PORT, default=8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT, default=30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT, default=30s"`
}

type DatabaseConfig struct {
	URL string `env:"DATABASE_URL, default=postgres://gateway:gateway@localhost:5432/gateway?sslmode=disable"`
}

type ProviderConfig struct {
	GroqAPIKey      string `env:"GROQ_API_KEY"`
	GeminiAPIKey    string `env:"GEMINI_API_KEY"`
	OpenAIAPIKey    string `env:"OPENAI_API_KEY"`
	AnthropicAPIKey string `env:"ANTHROPIC_API_KEY"`
}

type HealthCheckConfig struct {
	Interval         time.Duration `env:"HEALTH_CHECK_INTERVAL, default=30s"`
	Timeout          time.Duration `env:"HEALTH_CHECK_TIMEOUT, default=10s"`
	FailureThreshold int           `env:"HEALTH_CHECK_FAILURE_THRESHOLD, default=3"`
}

type LoggingConfig struct {
	Level  string `env:"LOG_LEVEL, default=debug"`
	Format string `env:"LOG_FORMAT, default=console"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		return Config{}, fmt.Errorf("process environment config: %w", err)
	}

	return cfg, nil
}

func (c Config) Addr() string {
	port := strings.TrimSpace(c.Server.Port)
	if port == "" {
		port = "8080"
	}

	if strings.HasPrefix(port, ":") {
		return port
	}

	return ":" + port
}
