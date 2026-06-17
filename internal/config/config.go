package config

import (
	"os"
	"strconv"

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
	_ = godotenv.Load()

	jwtExpireHours := getEnvInt("JWT_EXPIRE_HOURS", 24)
	otpTTLSeconds := getEnvInt("OTP_TTL_SECONDS", 300)
	maxTransfer := getEnvFloat("MAX_TRANSFER_AMOUNT", 10_000_000)
	dailyLimit := getEnvFloat("DAILY_LIMIT", 50_000_000)

	C = &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/qw_pay?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpireHours:    jwtExpireHours,
		OTPTTLSeconds:     otpTTLSeconds,
		MaxTransferAmount: maxTransfer,
		DailyLimit:        dailyLimit,
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

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
