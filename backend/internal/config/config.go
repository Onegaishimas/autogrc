// Package config provides configuration loading for the backend server.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the backend server.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Encryption  EncryptionConfig
	ServiceNow  ServiceNowConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// EncryptionConfig holds encryption key configuration.
type EncryptionConfig struct {
	Key string // Base64-encoded 32-byte AES-256 key
}

// ServiceNowConfig holds ServiceNow client configuration.
type ServiceNowConfig struct {
	Timeout    time.Duration
	MaxRetries int
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  time.Duration(getEnvInt("SERVER_READ_TIMEOUT_SECONDS", 30)) * time.Second,
			WriteTimeout: time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT_SECONDS", 30)) * time.Second,
			IdleTimeout:  time.Duration(getEnvInt("SERVER_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnvString("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnvString("DB_USER", "autogrc"),
			Password: getEnvString("DB_PASSWORD", ""),
			Name:     getEnvString("DB_NAME", "autogrc"),
			SSLMode:  getEnvString("DB_SSLMODE", "disable"),
		},
		Encryption: EncryptionConfig{
			Key: getEnvString("ENCRYPTION_KEY", ""),
		},
		ServiceNow: ServiceNowConfig{
			Timeout:    time.Duration(getEnvInt("SERVICENOW_TIMEOUT_SECONDS", 30)) * time.Second,
			MaxRetries: getEnvInt("SERVICENOW_MAX_RETRIES", 3),
		},
	}

	// Validate required configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return errors.New("DB_PASSWORD is required")
	}
	if c.Encryption.Key == "" {
		return errors.New("ENCRYPTION_KEY is required")
	}
	return nil
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// getEnvString gets a string environment variable or returns a default.
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable or returns a default.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
