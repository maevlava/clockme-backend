package config

import "os"

type Config struct {
	BackendPort string
}

func Load() *Config {
	backendPort := os.Getenv("BE_PORT")
	return &Config{
		BackendPort: backendPort,
	}
}
