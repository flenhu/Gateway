package config

import (
	"os"
	"strings"
)

type Config struct {
	Port string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{Port: port}
}

func (c Config) Addr() string {
	if strings.HasPrefix(c.Port, ":") {
		return c.Port
	}

	return ":" + c.Port
}
