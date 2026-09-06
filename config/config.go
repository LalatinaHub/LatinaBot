package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for LatinaBot.
type Config struct {
	AppEnv              string
	Port                string
	BotToken            string
	AdminID             int64
	GroupID             int64
	PromotionThreadID   int
	PromotionMessageID  int
	PublicNodeThreadID  int
	TursoDatabaseURL    string
	TursoAuthToken      string
	ServiceAccountURL   string
	GatewayPaymentKey   string
	TrakteerToken       string
	CloudflareEmail     string
	CloudflareAPIKey    string
	CloudflareAccountID string
	CloudflareZoneID    string
	WorkerServiceName   string
	WorkerEnvironment   string
	ServerPassword      string
	QuotaPerDonation    int64 // Quota added in MB per donation (default 1000000 MB = 1TB)
}

// LoadConfig loads environment variables from .env file (if present) and system environment.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", getEnv("WEBSITES_PORT", "8080")),
		BotToken:            os.Getenv("BOT_TOKEN"),
		TursoDatabaseURL:    os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:      os.Getenv("TURSO_AUTH_TOKEN"),
		ServiceAccountURL:   os.Getenv("SERVICE_ACCOUNT_URL"),
		GatewayPaymentKey:   os.Getenv("GATEWAY_PAYMENT_API_KEY"),
		TrakteerToken:       os.Getenv("TRAKTEER_TOKEN"),
		CloudflareEmail:     os.Getenv("CLOUDFLARE_EMAIL"),
		CloudflareAPIKey:    os.Getenv("CLOUDFLARE_API_KEY"),
		CloudflareAccountID: os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareZoneID:    os.Getenv("CLOUDFLARE_ZONE_ID"),
		WorkerServiceName:   os.Getenv("WORKER_SERVICE_NAME"),
		WorkerEnvironment:   os.Getenv("WORKER_ENVIRONMENT"),
		ServerPassword:      os.Getenv("SERVER_PASSWORD"),
		QuotaPerDonation:    1000000, // 1 TB in MB
	}

	// Parse integer values
	if adminIDStr := os.Getenv("ADMIN_ID"); adminIDStr != "" {
		val, err := strconv.ParseInt(adminIDStr, 10, 64)
		if err == nil {
			cfg.AdminID = val
		}
	}

	if groupIDStr := os.Getenv("GROUP_ID"); groupIDStr != "" {
		val, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err == nil {
			cfg.GroupID = val
		}
	}

	if promoThreadStr := os.Getenv("PROMOTION_THREAD_ID"); promoThreadStr != "" {
		val, err := strconv.Atoi(promoThreadStr)
		if err == nil {
			cfg.PromotionThreadID = val
		}
	}

	if promoMsgStr := os.Getenv("PROMOTION_MESSAGE_ID"); promoMsgStr != "" {
		val, err := strconv.Atoi(promoMsgStr)
		if err == nil {
			cfg.PromotionMessageID = val
		}
	}

	if publicThreadStr := os.Getenv("PUBLIC_NODE_THREAD_ID"); publicThreadStr != "" {
		val, err := strconv.Atoi(publicThreadStr)
		if err == nil {
			cfg.PublicNodeThreadID = val
		}
	}

	return cfg, nil
}

// Validate checks for critical configuration requirements.
func (c *Config) Validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN is required")
	}
	return nil
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "production" || strings.ToLower(c.AppEnv) == "prod"
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return !c.IsProduction()
}

// HTTPAddress returns address for internal HTTP server.
func (c *Config) HTTPAddress() string {
	if strings.HasPrefix(c.Port, ":") {
		return c.Port
	}
	return ":" + c.Port
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
