package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiryH  int
	CORSOrigin  string
	Env         string
}

func Load() Config {
	expiry := 24
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expiry = n
		}
	}

	port := getenv("PORT", "8080")
	return Config{
		Port:        port,
		DatabaseURL: getenv("DATABASE_URL", "postgres://helpdesk:helpdesk@localhost:5432/helpdesk?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiryH:  expiry,
		CORSOrigin:  getenv("CORS_ORIGIN", "http://localhost:5173"),
		Env:         getenv("APP_ENV", "development"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
