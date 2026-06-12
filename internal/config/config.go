package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения, загружаемые из переменных окружения.
type Config struct {
	DatabaseURL       string  // Строка подключения к PostgreSQL
	JWTSecret         string  // Секретный ключ для подписи JWT-токенов
	JWTExpireHours    int     // Время жизни JWT-токена в часах
	OTPTTLSeconds     int     // Время жизни OTP-кода в секундах
	MaxTransferAmount float64 // Максимальная сумма одного перевода
	DailyLimit        float64 // Дневной лимит переводов с одного счёта
	ServerPort        string  // Порт HTTP-сервера
}

// C — глобальный экземпляр конфигурации, доступный из любого пакета.
var C *Config

// Load загружает конфигурацию из .env файла и переменных окружения.
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
	}
}

// getEnv возвращает значение переменной окружения или fallback, если переменная не задана.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
