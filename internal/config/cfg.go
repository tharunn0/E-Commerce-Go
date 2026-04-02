package config

import (
	"os"
	"strconv"
)

type AppConfig struct {
	App      AppSettings
	Google   GoogleSettings
	SMTP     SMTPSettings
	Redis    RedisSettings
	Postgres PostgresSettings
	Security SecuritySettings
	Razorpay RazorpaySettings
	Order    OrderSettings
	Cart     CartSettings
}

type AppSettings struct {
	Env  string
	Host string
	Port string
}

type GoogleSettings struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type SMTPSettings struct {
	Host string
	Port int
	User string
	Pass string
}

type RedisSettings struct {
	URL      string
	Password string
	DB       int
}

type PostgresSettings struct {
	User     string
	Password string
	Host     string
	Port     string
	DB       string
	SSL      string
}

type SecuritySettings struct {
	JWTSecret               string
	JWTExpiryMinutes        int
	OTPExpiryMinutes        int
	PasswordResetExpiry     int
	EmailVerificationExpiry int
}

type RazorpaySettings struct {
	KeyID         string
	KeySecret     string
	WebhookSecret string
}

type OrderSettings struct {
	MaxOrderAmount    int64
	MaxCodOrderAmount int64
	ShippingCost      int64
	DeliveryTime      int64
}

type CartSettings struct {
	MaxQuantity int64
}

func LoadConfig() *AppConfig {
	return &AppConfig{
		App: AppSettings{
			Env:  getEnv("APP_ENV", "development"),
			Host: getEnv("APP_HOST", "0.0.0.0"),
			Port: getEnv("APP_PORT", "8080"),
		},

		Google: GoogleSettings{
			ClientID:     getEnv("GOOGLE_AUTH_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_AUTH_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_AUTH_REDIRECT_URL", ""),
		},

		SMTP: SMTPSettings{
			Host: getEnv("SMTP_HOST", ""),
			Port: getEnvAsInt("SMTP_PORT", 587),
			User: getEnv("SMTP_USER", ""),
			Pass: getEnv("SMTP_PASSWORD", ""),
		},

		Redis: RedisSettings{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},

		Postgres: PostgresSettings{
			User:     getEnv("POSTGRES_USER", ""),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			DB:       getEnv("POSTGRES_DB", ""),
			SSL:      getEnv("POSTGRES_SSL", "disable"),
		},

		Security: SecuritySettings{
			JWTSecret:               getEnv("JWT_SECRET", ""),
			JWTExpiryMinutes:        getEnvAsInt("JWT_EXPIRY", 60),
			OTPExpiryMinutes:        getEnvAsInt("OTP_EXPIRY", 5),
			PasswordResetExpiry:     getEnvAsInt("PASSWORD_RESET_EXPIRY", 15),
			EmailVerificationExpiry: getEnvAsInt("EMAIL_VERIFICATION_EXPIRY", 15),
		},

		Razorpay: RazorpaySettings{
			KeyID:         getEnv("RAZORPAY_ID", ""),
			KeySecret:     getEnv("RAZORPAY_SECRET", ""),
			WebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
		},
		Order: OrderSettings{
			MaxOrderAmount:    int64(getEnvAsInt("ALL_ORDER_MAX_AMOUNT", 10_00_000)),
			MaxCodOrderAmount: int64(getEnvAsInt("COD_ORDER_MAX_AMOUNT", 50_000)),
			ShippingCost:      int64(getEnvAsInt("SHIPPING_COST", 150)),
			DeliveryTime:      int64(getEnvAsInt("DELIVERY_TIME", 7)),
		},
		Cart: CartSettings{
			MaxQuantity: int64(getEnvAsInt("CART_MAX_QUANTITY", 5)),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}
