package utils

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPTimeout      = 30 * time.Second
	defaultToolTimeout      = 60 * time.Second
	defaultSlowRequest      = 2 * time.Second
	defaultMaxPerPage       = 100
	defaultMaxResponseBytes = 1 << 20
)

func HTTPTimeout() time.Duration {
	return envDuration("GITEE_ENT_HTTP_TIMEOUT", defaultHTTPTimeout)
}

func ToolTimeout() time.Duration {
	return envDuration("GITEE_ENT_TOOL_TIMEOUT", defaultToolTimeout)
}

func SlowRequestThreshold() time.Duration {
	return envDuration("GITEE_ENT_SLOW_REQUEST_MS", defaultSlowRequest, time.Millisecond)
}

func MaxPerPage() int {
	return envInt("GITEE_ENT_MAX_PER_PAGE", defaultMaxPerPage)
}

func MaxResponseBytes() int {
	return envInt("GITEE_ENT_MAX_RESPONSE_BYTES", defaultMaxResponseBytes)
}

func envDuration(name string, fallback time.Duration, unit ...time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	if len(unit) > 0 {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return fallback
		}
		return time.Duration(parsed) * unit[0]
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
