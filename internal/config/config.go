package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName             string
	Port                string
	HTTPTimeout         time.Duration
	GoogleMapsUserAgent string
	GoogleMapsCookie    string
}

func Load() Config {
	// Dotenv config
	return Config{
		AppName:             getEnv("APP_NAME", "Rasanta"),
		Port:                getEnv("PORT", "3000"),
		HTTPTimeout:         time.Duration(getEnvInt("HTTP_TIMEOUT_SECONDS", 30)) * time.Second,
		GoogleMapsUserAgent: getEnv("GOOGLE_MAPS_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"),
		GoogleMapsCookie:    os.Getenv("GOOGLE_MAPS_COOKIE"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		// Fallback ke default kalau parsing gagal
		return fallback
	}
	return value
}
