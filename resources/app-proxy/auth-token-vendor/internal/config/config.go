package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type ctxConfigKey int

// RequestIDKey is the key that holds the unique request ID in a request context.
const ConfigKey ctxConfigKey = 0

type (
	// Config provides the system configuration.
	Config struct {
		Server     Server
		Logging    Logging
		Databricks Databricks
	}

	// Logging provides the logging configuration.
	Logging struct {
		LogLevel string `envconfig:"LOG_LEVEL"`
	}

	// Server provides the server configuration.
	Server struct {
		Host      string `envconfig:"SERVER_HOST" default:"localhost:7080"`
		ShutTime  int    `envconfig:"SERVER_SHUTTIME" default:"5"`
		ReadTime  int    `envconfig:"SERVER_READTIME" default:"5"`
		WriteTime int    `envconfig:"SERVER_WRITETIME" default:"5"`
		CacheTime int    `envconfig:"SERVER_CACHETIME" default:"3600"`
	}

	// Server provides the server configuration.
	Databricks struct {
		Hostname     string `envconfig:"DATABRICKS_WORKSPACE_URI" default:"http://localhost:7200"`
		ClientId     string `envconfig:"DATABRICKS_CLIENT_ID" default:"unset"`
		ClientSecret string `envconfig:"DATABRICKS_CLIENT_SECRET" default:"unset"`
		ReadTime     int    `envconfig:"DATABRICKS_READTIME" default:"30"`
	}
)

// Environ returns the settings from the environment.
func Environ() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	return cfg, err
}

// String returns the configuration in string format.
func (c *Config) String() string {
	out, _ := yaml.Marshal(c)
	return string(out)
}
