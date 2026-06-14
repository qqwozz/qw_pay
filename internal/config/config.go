package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTExpireHours    int
	OTPTTLSeconds     int
	MaxTransferAmount float64
	DailyLimit        float64
	ServerPort        string
	RedisAddr         string
}

var C *Config

func Load() {
	godotenv.Load()
	C = &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpireHours:    24,
		OTPTTLSeconds:     300,
		MaxTransferAmount: 10_000_000,
		DailyLimit:        50_000_000,
		ServerPort:        getEnv("PORT", "8080"),
		RedisAddr:         getEnv("REDIS_ADDR", "127.0.0.1:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
