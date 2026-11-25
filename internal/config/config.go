package config

import (
	"log"
	"os"

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

	return &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
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
