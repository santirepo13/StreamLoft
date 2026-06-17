package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseHost      string
	DatabasePort      int
	DatabaseUser      string
	DatabasePassword  string
	DatabaseName      string
	APIPort           string
	SRSURL            string
	RTMPURL           string
	JWTSecret         string
	LogLevel          string
	EncryptionKeyPath string
	EnvFilePath       string
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		c.DatabaseUser, c.DatabasePassword,
		c.DatabaseHost, c.DatabasePort,
		c.DatabaseName)
}

func New() (*Config, error) {
	cfg := &Config{
		EnvFilePath:       "/opt/StreamLoft/api.env",
		EncryptionKeyPath: "/opt/StreamLoft/streamloft_master.key",
		DatabasePort:      5432,
		DatabaseName:      "streamloft",
		APIPort:           "10002",
		RTMPURL:           "rtmp://74.81.35.20:10004/live",
		SRSURL:            "http://74.81.35.20:10003",
		LogLevel:          "info",
	}

	if err := cfg.loadFromEnvFile(); err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", cfg.EnvFilePath, err)
	}

	cfg.loadFromEnvironment()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadFromEnvFile() error {
	if _, err := os.Stat(c.EnvFilePath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s (create from .env.example)", c.EnvFilePath)
	}

	if err := godotenv.Load(c.EnvFilePath); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}

	return nil
}

func (c *Config) loadFromEnvironment() {
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		c.DatabaseHost = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.DatabasePort = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		c.DatabaseUser = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		c.DatabasePassword = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		c.DatabaseName = v
	}
	if v := os.Getenv("API_PORT"); v != "" {
		c.APIPort = v
	}
	if v := os.Getenv("SRS_URL"); v != "" {
		c.SRSURL = v
	}
	if v := os.Getenv("RTMP_URL"); v != "" {
		c.RTMPURL = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("ENCRYPTION_KEY_PATH"); v != "" {
		c.EncryptionKeyPath = v
	}
}

func (c *Config) validate() error {
	var missing []string

	if c.DatabaseHost == "" {
		missing = append(missing, "DATABASE_HOST")
	}
	if c.DatabaseUser == "" {
		missing = append(missing, "DATABASE_USER")
	}
	if c.DatabasePassword == "" {
		missing = append(missing, "DATABASE_PASSWORD")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}

	if _, err := os.Stat(c.EncryptionKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("encryption key file not found: %s", c.EncryptionKeyPath)
	}

	return nil
}

func (c *Config) TokenExpiry() time.Time {
	return time.Now().Add(72 * time.Hour)
}

func (c *Config) Getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
