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
	Creator  CreatorConfig
	Agent    AgentConfig
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	SSLMode        string
	MaxConnections int
	MaxIdleConns   int
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
	Backend       string
	Dir           string
	MaxSizeMB     int
	AllowedExt    []string
	GCSBucket     string
	GCSPrefix     string
	GCSPublicBase string
}

type AdminConfig struct {
	Email    string
	Password string
}

type CreatorConfig struct {
	Email    string
	Password string
}

type AgentConfig struct {
	Token        string
	CreatorEmail string
}

func Load() (*Config, error) {
	// Load .env file if exists (development)
	// Try current directory first, then parent directory
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../.env")
	}

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		Env:     getEnv("ENV", "development"),
		BaseURL: getEnv("BASE_URL", "http://localhost:8080"),
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "5432"),
			User:           getEnv("DB_USER", "zenbali"),
			Password:       getEnv("DB_PASSWORD", "zenbali_dev_password"),
			Name:           getEnv("DB_NAME", "zenbali"),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
			MaxConnections: getEnvInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:   getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "default-dev-secret-change-in-production-min-32-chars"),
			ExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
		},
		Stripe: StripeConfig{
			SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
			PublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
			WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
			PriceCents:     int64(getEnvInt("STRIPE_PRICE_CENTS", 100)),
		},
		Upload: UploadConfig{
			Backend:       getEnv("UPLOAD_BACKEND", "local"),
			Dir:           getEnv("UPLOAD_DIR", "./uploads"),
			MaxSizeMB:     getEnvInt("MAX_UPLOAD_SIZE_MB", 5),
			AllowedExt:    []string{".jpg", ".jpeg", ".png", ".webp"},
			GCSBucket:     getEnv("GCS_BUCKET", ""),
			GCSPrefix:     getEnv("GCS_PREFIX", ""),
			GCSPublicBase: getEnv("GCS_PUBLIC_BASE_URL", ""),
		},
		Admin: AdminConfig{
			Email:    getEnv("ADMIN_EMAIL", "admin@zenbali.org"),
			Password: getEnv("ADMIN_PASSWORD", "Teameditor@123"),
		},
		Creator: CreatorConfig{
			Email:    getEnv("CREATOR_EMAIL", "creator@zenbali.org"),
			Password: getEnv("CREATOR_PASSWORD", "admin123"),
		},
		Agent: AgentConfig{
			Token:        getEnv("AGENT_API_TOKEN", ""),
			CreatorEmail: getEnv("AGENT_CREATOR_EMAIL", getEnv("CREATOR_EMAIL", "creator@zenbali.org")),
		},
	}

	if cfg.Upload.Backend == "gcs" && cfg.Upload.GCSBucket == "" {
		return nil, fmt.Errorf("GCS_BUCKET is required when UPLOAD_BACKEND=gcs")
	}

	// Create upload directory if not exists for local storage.
	if cfg.Upload.Backend != "gcs" {
		if err := os.MkdirAll(cfg.Upload.Dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create upload directory: %w", err)
		}
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
