package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	StripeTestKey       string
	StripeWebhookSecret string
	StripeReturnURL     string
}

func Load() *Config {
	// Загружаем энву
	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: файл .env не найден, используем системные переменные")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if len(strings.TrimSpace(jwtSecret)) < 32 || jwtSecret == "super_secret_key_change_me_in_production" {
		log.Fatal("JWT_SECRET должен быть случайным секретом длиной не менее 32 байт")
	}

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		JWTSecret:           jwtSecret,
		StripeTestKey:       getEnv("STRIPE_TEST_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeReturnURL:     getEnv("STRIPE_RETURN_URL", "http://localhost:3000/bookings/payment-return"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
