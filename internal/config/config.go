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
	SingBoxContainerName      string
	SingBoxAutoRestart        bool
	SingBoxRestartDriver      string
	SingBoxDockerSocket       string
	SingBoxRestartCommand     string
	SingBoxRestartArgs        []string
	SingBoxRestartTimeout     time.Duration
	SingBoxV2RayAPIAddr       string
	SingBoxV2RayStatsPattern  string
	SingBoxV2RayAPITimeout    time.Duration
	SourceSyncPollInterval    time.Duration
	SourceSyncBatchLimit      int
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
		SingBoxContainerName:      env("SING_BOX_CONTAINER_NAME", "fluxgate-sing-box"),
		SingBoxAutoRestart:        envBool("SING_BOX_AUTO_RESTART", false),
		SingBoxRestartDriver:      strings.ToLower(env("SING_BOX_RESTART_DRIVER", "docker")),
		SingBoxDockerSocket:       env("SING_BOX_DOCKER_SOCKET", "/var/run/docker.sock"),
		SingBoxRestartCommand:     env("SING_BOX_RESTART_COMMAND", ""),
		SingBoxRestartArgs:        envFields("SING_BOX_RESTART_ARGS"),
		SingBoxRestartTimeout:     time.Duration(envInt("SING_BOX_RESTART_TIMEOUT_SECONDS", 15)) * time.Second,
		SingBoxV2RayAPIAddr:       env("SING_BOX_V2RAY_API_ADDR", ""),
		SingBoxV2RayStatsPattern:  env("SING_BOX_V2RAY_STATS_PATTERN", "traffic"),
		SingBoxV2RayAPITimeout:    time.Duration(envInt("SING_BOX_V2RAY_API_TIMEOUT_SECONDS", 5)) * time.Second,
		SourceSyncPollInterval:    time.Duration(envInt("SOURCE_SYNC_POLL_SECONDS", 60)) * time.Second,
		SourceSyncBatchLimit:      envInt("SOURCE_SYNC_BATCH_LIMIT", 20),
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

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
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

func envFields(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	if strings.Contains(raw, ",") {
		var result []string
		for _, item := range strings.Split(raw, ",") {
			if value := strings.TrimSpace(item); value != "" {
				result = append(result, value)
			}
		}
		return result
	}
	return strings.Fields(raw)
}
