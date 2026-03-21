package config

// =============================================================================
// 📖 LEARNING NOTES — Config Loader (Step 3 — MySQL)
// =============================================================================
//
// Config now includes DatabaseConfig for MySQL connection settings.
// DSN format for MySQL: user:password@tcp(host:port)/dbname?parseTime=true
// =============================================================================

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application settings.
type Config struct {
	Port string // PORT — which port to listen on
	Env  string // APP_ENV — "development" or "production"

	// Database settings (Step 3)
	DB DatabaseConfig

	// JWT settings (Step 4)
	JWT JWTConfig
}

// DatabaseConfig holds MySQL connection settings.
type DatabaseConfig struct {
	Host     string // DB_HOST
	Port     string // DB_PORT
	User     string // DB_USER
	Password string // DB_PASSWORD
	Name     string // DB_NAME
}

// JWTConfig holds authentication settings.
type JWTConfig struct {
	Secret string
	TTL    time.Duration
}


// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	// Server defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_ENV", "development")

	// Database defaults (match docker-compose.yml)
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("DB_USER", "streaming")
	viper.SetDefault("DB_PASSWORD", "secret")
	viper.SetDefault("DB_NAME", "streaming")

	// Read config
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("⚠️  No .env file found, using defaults/env vars: %v\n", err)
	}

	jwtTTL, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TTL"))
	if err != nil {
		jwtTTL = 15 * time.Minute // Default fallback
	}

	return &Config{
		Port: viper.GetString("PORT"),
		Env:  viper.GetString("APP_ENV"),
		DB: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			TTL:    jwtTTL,
		},
	}, nil
}
