package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Env      string
	BaseURL  string
	Database DatabaseConfig
	JWT      JWTConfig
	Stripe   StripeConfig
	Upload   UploadConfig
	Admin    AdminConfig
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type StripeConfig struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
	PriceCents     int64
}

type UploadConfig struct {
	Dir        string
	MaxSizeMB  int
	AllowedExt []string
}

type AdminConfig struct {
	Email    string
	Password string
}

func Load() (*Config, error) {
	// Load .env file if exists (development)
	_ = godotenv.Load()

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		Env:     getEnv("ENV", "development"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "zenbali"),
			Password:        getEnv("DB_PASSWORD", "zenbali_dev_password"),
			Name:            getEnv("DB_NAME", "zenbali"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:  getEnvInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "default-dev-secret-change-in-production-min-32-chars"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
		},
		Stripe: StripeConfig{
			SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
			PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
			PriceCents:     int64(getEnvInt("STRIPE_PRICE_CENTS", 1000)),
		},
		Upload: UploadConfig{
			Dir:        getEnv("UPLOAD_DIR", "./uploads"),
			MaxSizeMB:  getEnvInt("MAX_UPLOAD_SIZE_MB", 5),
			AllowedExt: []string{".jpg", ".jpeg", ".png", ".webp"},
		},
		Admin: AdminConfig{
			Email:    getEnv("ADMIN_EMAIL", "admin@zenbali.org"),
			Password: getEnv("ADMIN_PASSWORD", "admin123"),
		},
	}

	// Create upload directory if not exists
	if err := os.MkdirAll(cfg.Upload.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return cfg, nil
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
