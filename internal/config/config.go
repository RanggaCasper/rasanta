package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var loadEnvOnce sync.Once

type Config struct {
	AppName             string
	Port                string
	HTTPTimeout         time.Duration
	GoogleMapsUserAgent string
	GoogleMapsCookie    string
}

func Load() Config {
	loadEnvOnce.Do(loadDotEnvIfExists)

	return Config{
		AppName:             getEnv("APP_NAME", "Rasanta"),
		Port:                getEnv("PORT", "8000"),
		HTTPTimeout:         time.Duration(getEnvInt("HTTP_TIMEOUT_SECONDS", 30)) * time.Second,
		GoogleMapsUserAgent: getEnv("GOOGLE_MAPS_USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"),
		GoogleMapsCookie:    loadCookie(),
	}
}

func loadDotEnvIfExists() {
	content, err := os.ReadFile(".env")
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}

		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		_ = os.Setenv(key, value)
	}
}

func loadCookie() string {
	cookieFile := getEnv("GOOGLE_MAPS_COOKIE_FILE", "")
	if cookieFile != "" {
		if cookie := readCookieFile(cookieFile); cookie != "" {
			return cookie
		}
	}

	return strings.TrimSpace(os.Getenv("GOOGLE_MAPS_COOKIE"))
}

func readCookieFile(cookieFile string) string {
	candidates := []string{cookieFile}
	if !filepath.IsAbs(cookieFile) {
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(wd, cookieFile))
		}
		if exePath, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exePath)
			candidates = append(candidates, filepath.Join(exeDir, cookieFile))
			candidates = append(candidates, filepath.Join(filepath.Dir(exeDir), cookieFile))
		}
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		cleanPath := filepath.Clean(candidate)
		if _, exists := seen[cleanPath]; exists {
			continue
		}
		seen[cleanPath] = struct{}{}

		data, err := os.ReadFile(cleanPath)
		if err != nil {
			continue
		}

		cookie := strings.TrimSpace(string(data))
		if cookie != "" {
			return cookie
		}
	}

	return ""
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
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
