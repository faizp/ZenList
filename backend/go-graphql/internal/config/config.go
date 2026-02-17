package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contains runtime configuration for the API process.
type Config struct {
	AppEnv             string
	HTTPPort           string
	DatabaseURL        string
	DefaultUserName    string
	DefaultUserEmail   string
	DefaultUserTZ      string
	DefaultUserAvatar  string
	RequestTimeout     time.Duration
	QueryTimeout       time.Duration
	DBMaxConns         int32
	DBMinConns         int32
	DBHealthCheckEvery time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DefaultUserName:    getEnv("DEFAULT_USER_NAME", "ZenList User"),
		DefaultUserEmail:   getEnv("DEFAULT_USER_EMAIL", "user@zenlist.local"),
		DefaultUserTZ:      getEnv("DEFAULT_USER_TIMEZONE", "UTC"),
		DefaultUserAvatar:  getEnv("DEFAULT_USER_AVATAR_URL", ""),
		RequestTimeout:     getDuration("REQUEST_TIMEOUT", 20*time.Second),
		QueryTimeout:       getDuration("QUERY_TIMEOUT", 3*time.Second),
		DBMaxConns:         int32(getInt("DB_MAX_CONNS", 20)),
		DBMinConns:         int32(getInt("DB_MIN_CONNS", 2)),
		DBHealthCheckEvery: getDuration("DB_HEALTHCHECK_PERIOD", 30*time.Second),
	}

	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if strings.TrimSpace(cfg.DefaultUserEmail) == "" {
		return Config{}, errors.New("DEFAULT_USER_EMAIL is required")
	}
	if strings.TrimSpace(cfg.DefaultUserName) == "" {
		return Config{}, errors.New("DEFAULT_USER_NAME is required")
	}
	if strings.TrimSpace(cfg.DefaultUserTZ) == "" {
		return Config{}, errors.New("DEFAULT_USER_TIMEZONE is required")
	}
	if cfg.RequestTimeout <= 0 {
		return Config{}, errors.New("REQUEST_TIMEOUT must be positive")
	}
	if cfg.QueryTimeout <= 0 {
		return Config{}, errors.New("QUERY_TIMEOUT must be positive")
	}
	if cfg.DBMaxConns < 1 {
		return Config{}, errors.New("DB_MAX_CONNS must be >= 1")
	}
	if cfg.DBMinConns < 0 || cfg.DBMinConns > cfg.DBMaxConns {
		return Config{}, errors.New("DB_MIN_CONNS must be between 0 and DB_MAX_CONNS")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
