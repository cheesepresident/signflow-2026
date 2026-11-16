package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	Port           string
	Env            string
	BaseURL        string
	SessionSecret  string
	EmailSender    string
	ResendAPIKey   string
	EmailFrom      string
	UploadDir      string
	MaxUploadBytes int64
}

func (c Config) IsProd() bool { return c.Env == "prod" }

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:    getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/signflow?sslmode=disable"),
		Port:           getenv("PORT", "8080"),
		Env:            getenv("APP_ENV", "dev"),
		BaseURL:        os.Getenv("BASE_URL"),
		SessionSecret:  os.Getenv("SESSION_SECRET"),
		EmailSender:    getenv("EMAIL_SENDER", "console"),
		ResendAPIKey:   os.Getenv("RESEND_API_KEY"),
		EmailFrom:      os.Getenv("EMAIL_FROM"),
		UploadDir:      getenv("UPLOAD_DIR", "uploads"),
		MaxUploadBytes: 25 << 20,
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	if cfg.IsProd() && os.Getenv("BASE_URL") == "" {
		return Config{}, fmt.Errorf("BASE_URL is required in production")
	}

	if cfg.IsProd() && cfg.SessionSecret == "" {
		return Config{}, fmt.Errorf("SESSION_SECRET is required in production")
	}
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = "dev-insecure-session-secret-change-me"
	}

	return cfg, nil
}

func getenv(key, fallback string) string {

	if v := os.Getenv(key); v != "" {

		return v

	}

	return fallback

}
