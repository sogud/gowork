// Package server provides HTTP API server for the multi-agent workflow system.
package server

import (
	"time"
)

// Config holds the API server configuration.
type Config struct {
	// Host is the server host address.
	// Default: "0.0.0.0"
	Host string

	// Port is the server port.
	// Default: 8080
	Port int

	// ReadTimeout is the maximum duration for reading the entire request.
	// Default: 15 seconds
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// Default: 15 seconds
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request.
	// Default: 60 seconds
	IdleTimeout time.Duration

	// ShutdownTimeout is the maximum duration to wait for active connections to close during shutdown.
	// Default: 30 seconds
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Option is a function that modifies the Config.
type Option func(*Config)

// WithHost sets the server host.
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort sets the server port.
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ShutdownTimeout = timeout
	}
}

// Apply applies all options to the config.
func (c *Config) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}

// Validate checks if the configuration is valid.
// Returns an error if any required field is invalid.
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return &ConfigError{Field: "Port", Message: "port must be between 1 and 65535"}
	}

	if c.Host == "" {
		return &ConfigError{Field: "Host", Message: "host cannot be empty"}
	}

	if c.ReadTimeout <= 0 {
		return &ConfigError{Field: "ReadTimeout", Message: "read timeout must be positive"}
	}

	if c.WriteTimeout <= 0 {
		return &ConfigError{Field: "WriteTimeout", Message: "write timeout must be positive"}
	}

	if c.IdleTimeout <= 0 {
		return &ConfigError{Field: "IdleTimeout", Message: "idle timeout must be positive"}
	}

	if c.ShutdownTimeout <= 0 {
		return &ConfigError{Field: "ShutdownTimeout", Message: "shutdown timeout must be positive"}
	}

	return nil
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " - " + e.Message
}