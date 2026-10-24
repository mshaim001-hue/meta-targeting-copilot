package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL    string
	HTTPAddr       string
	FrontendOrigin string
	JWTSecret      string

	LLMAPIKey  string
	LLMBaseURL string
	LLMModel   string

	MetaAppID       string
	MetaAppSecret   string
	MetaAccessToken string
	MetaAdAccountID string
	MetaAPIVersion  string

	CacheTTL time.Duration
}

func Load() Config {
	ttlDays := envInt("TARGETING_CACHE_TTL_DAYS", 14)
	return Config{
		DatabaseURL:     env("DATABASE_URL", "postgres://copilot:copilot@localhost:5432/copilot?sslmode=disable"),
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		FrontendOrigin:  env("FRONTEND_ORIGIN", "http://localhost:5173"),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
		LLMAPIKey:       env("LLM_API_KEY", ""),
		LLMBaseURL:      env("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMModel:        env("LLM_MODEL", "gpt-4.1-mini"),
		MetaAppID:       env("META_APP_ID", ""),
		MetaAppSecret:   env("META_APP_SECRET", ""),
		MetaAccessToken: env("META_ACCESS_TOKEN", ""),
		MetaAdAccountID: env("META_AD_ACCOUNT_ID", ""),
		MetaAPIVersion:  env("META_API_VERSION", "v26.0"),
		CacheTTL:        time.Duration(ttlDays) * 24 * time.Hour,
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
