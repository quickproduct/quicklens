// Package config loads runtime configuration via Viper (env vars + optional
// config file) into a typed Config.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config is the resolved application configuration.
type Config struct {
	Env  string
	Port string

	DatabasePath string

	JWTSecret string

	AdminEmail    string
	AdminPassword string
}

// Load reads configuration via Viper: defaults, then environment variables,
// then an optional config.yaml. Environment names match the historical keys
// (PORT, DB_PATH/SQLITE_DB_PATH, JWT_SECRET_KEY, ...).
func Load() Config {
	v := viper.New()

	v.SetDefault("environment", "development")
	v.SetDefault("port", "8000")
	v.SetDefault("db_path", "")
	v.SetDefault("sqlite_db_path", "quicklens.db")
	v.SetDefault("jwt_secret_key", "quicklens-default-secret-change-me-in-production")
	v.SetDefault("default_admin_email", "admin@quicklens.dev")
	v.SetDefault("default_admin_password", "admin123")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/quicklens")
	_ = v.ReadInConfig() // optional; ignore not-found

	dbPath := v.GetString("db_path")
	if dbPath == "" {
		dbPath = v.GetString("sqlite_db_path")
	}

	return Config{
		Env:           v.GetString("environment"),
		Port:          v.GetString("port"),
		DatabasePath:  dbPath,
		JWTSecret:     v.GetString("jwt_secret_key"),
		AdminEmail:    v.GetString("default_admin_email"),
		AdminPassword: v.GetString("default_admin_password"),
	}
}
