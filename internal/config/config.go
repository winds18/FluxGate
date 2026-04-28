package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                    string
	HTTPAddr                  string
	PublicBaseURL             string
	DBPath                    string
	LogDir                    string
	TokenSecret               string
	SessionSecret             string
	AdminBootstrapUsername    string
	AdminBootstrapPassword    string
	GatewayHost               string
	DefaultVLESSPort          int
	SingBoxConfigPath         string
	SingBoxPreviousConfigPath string
	StatsPollInterval         time.Duration
	Version                   string
}

func Load() Config {
	return Config{
		AppEnv:                    env("APP_ENV", "development"),
		HTTPAddr:                  env("HTTP_ADDR", "127.0.0.1:8080"),
		PublicBaseURL:             strings.TrimRight(env("PUBLIC_BASE_URL", "http://127.0.0.1:8080"), "/"),
		DBPath:                    env("DB_PATH", "data/fluxgate.db"),
		LogDir:                    env("LOG_DIR", "logs/fluxgate"),
		TokenSecret:               env("TOKEN_SECRET", "dev-token-secret-change-me"),
		SessionSecret:             env("SESSION_SECRET", "dev-session-secret-change-me"),
		AdminBootstrapUsername:    env("ADMIN_BOOTSTRAP_USERNAME", ""),
		AdminBootstrapPassword:    env("ADMIN_BOOTSTRAP_PASSWORD", ""),
		GatewayHost:               env("GATEWAY_HOST", "127.0.0.1"),
		DefaultVLESSPort:          envInt("DEFAULT_VLESS_PORT", 8443),
		SingBoxConfigPath:         env("SING_BOX_CONFIG_PATH", "data/sing-box/config.json"),
		SingBoxPreviousConfigPath: env("SING_BOX_PREVIOUS_CONFIG_PATH", "data/sing-box/config.previous.json"),
		StatsPollInterval:         time.Duration(envInt("STATS_POLL_INTERVAL_SECONDS", 30)) * time.Second,
		Version:                   env("APP_VERSION", "dev"),
	}
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
