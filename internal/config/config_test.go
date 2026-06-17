package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Setenv("JWT_SECRET", "test-secret-for-defaults")
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("REDIS_ADDR")
	_ = os.Unsetenv("JWT_EXPIRE_HOURS")
	_ = os.Unsetenv("OTP_TTL_SECONDS")
	_ = os.Unsetenv("MAX_TRANSFER_AMOUNT")
	_ = os.Unsetenv("DAILY_LIMIT")
	defer func() { _ = os.Unsetenv("JWT_SECRET") }()

	Load()

	if C == nil {
		t.Fatal("config not loaded")
	}
	if C.DatabaseURL == "" {
		t.Error("DatabaseURL should not be empty")
	}
	if C.JWTSecret != "test-secret-for-defaults" {
		t.Errorf("expected JWT secret test-secret-for-defaults, got %s", C.JWTSecret)
	}
	if C.ServerPort != "8080" {
		t.Errorf("expected port 8080, got %s", C.ServerPort)
	}
	if C.RedisAddr != "127.0.0.1:6379" {
		t.Errorf("expected redis addr 127.0.0.1:6379, got %s", C.RedisAddr)
	}
	if C.JWTExpireHours != 24 {
		t.Errorf("expected JWT expire 24, got %d", C.JWTExpireHours)
	}
	if C.OTPTTLSeconds != 300 {
		t.Errorf("expected OTP TTL 300, got %d", C.OTPTTLSeconds)
	}
	if C.MaxTransferAmount != 10_000_000 {
		t.Errorf("expected max transfer 10000000, got %f", C.MaxTransferAmount)
	}
	if C.DailyLimit != 50_000_000 {
		t.Errorf("expected daily limit 50000000, got %f", C.DailyLimit)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	_ = os.Setenv("PORT", "9999")
	_ = os.Setenv("REDIS_ADDR", "redis:6379")
	_ = os.Setenv("JWT_SECRET", "super-secret-key-for-testing")
	_ = os.Setenv("JWT_EXPIRE_HOURS", "48")
	_ = os.Setenv("OTP_TTL_SECONDS", "600")
	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("REDIS_ADDR")
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("JWT_EXPIRE_HOURS")
		_ = os.Unsetenv("OTP_TTL_SECONDS")
	}()

	Load()

	if C.ServerPort != "9999" {
		t.Errorf("expected port 9999, got %s", C.ServerPort)
	}
	if C.RedisAddr != "redis:6379" {
		t.Errorf("expected redis addr redis:6379, got %s", C.RedisAddr)
	}
	if C.JWTSecret != "super-secret-key-for-testing" {
		t.Errorf("expected JWT secret super-secret-key-for-testing, got %s", C.JWTSecret)
	}
	if C.JWTExpireHours != 48 {
		t.Errorf("expected JWT expire 48, got %d", C.JWTExpireHours)
	}
	if C.OTPTTLSeconds != 600 {
		t.Errorf("expected OTP TTL 600, got %d", C.OTPTTLSeconds)
	}
}

func TestLoad_ValidationPanics(t *testing.T) {
	_ = os.Setenv("JWT_SECRET", "")
	defer func() {
		_ = os.Unsetenv("JWT_SECRET")
	}()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty JWT_SECRET")
		}
	}()

	Load()
}

func TestGetEnv(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		_ = os.Setenv("TEST_GETENV_KEY", "hello")
		defer func() { _ = os.Unsetenv("TEST_GETENV_KEY") }()
		result := getEnv("TEST_GETENV_KEY", "default")
		if result != "hello" {
			t.Errorf("expected 'hello', got '%s'", result)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		_ = os.Unsetenv("TEST_GETENV_KEY2")
		result := getEnv("TEST_GETENV_KEY2", "fallback")
		if result != "fallback" {
			t.Errorf("expected 'fallback', got '%s'", result)
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("returns parsed int", func(t *testing.T) {
		_ = os.Setenv("TEST_INT_KEY", "42")
		defer func() { _ = os.Unsetenv("TEST_INT_KEY") }()
		result := getEnvInt("TEST_INT_KEY", 0)
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("returns fallback for invalid int", func(t *testing.T) {
		_ = os.Setenv("TEST_INT_KEY2", "not-a-number")
		defer func() { _ = os.Unsetenv("TEST_INT_KEY2") }()
		result := getEnvInt("TEST_INT_KEY2", 99)
		if result != 99 {
			t.Errorf("expected 99, got %d", result)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		_ = os.Unsetenv("TEST_INT_KEY3")
		result := getEnvInt("TEST_INT_KEY3", 77)
		if result != 77 {
			t.Errorf("expected 77, got %d", result)
		}
	})
}

func TestGetEnvFloat(t *testing.T) {
	t.Run("returns parsed float", func(t *testing.T) {
		_ = os.Setenv("TEST_FLOAT_KEY", "3.14")
		defer func() { _ = os.Unsetenv("TEST_FLOAT_KEY") }()
		result := getEnvFloat("TEST_FLOAT_KEY", 0)
		if result != 3.14 {
			t.Errorf("expected 3.14, got %f", result)
		}
	})

	t.Run("returns fallback for invalid float", func(t *testing.T) {
		_ = os.Setenv("TEST_FLOAT_KEY2", "not-a-float")
		defer func() { _ = os.Unsetenv("TEST_FLOAT_KEY2") }()
		result := getEnvFloat("TEST_FLOAT_KEY2", 1.0)
		if result != 1.0 {
			t.Errorf("expected 1.0, got %f", result)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		_ = os.Unsetenv("TEST_FLOAT_KEY3")
		result := getEnvFloat("TEST_FLOAT_KEY3", 2.5)
		if result != 2.5 {
			t.Errorf("expected 2.5, got %f", result)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	validCfg := &Config{
		DatabaseURL:       "postgres://localhost/db",
		JWTSecret:         "secret",
		JWTExpireHours:    24,
		OTPTTLSeconds:     300,
		MaxTransferAmount: 10000000,
		DailyLimit:        50000000,
		ServerPort:        "8080",
		RedisAddr:         "localhost:6379",
	}
	if err := validCfg.validate(); err != nil {
		t.Errorf("valid config should not error: %v", err)
	}

	invalidCfg := &Config{}
	if err := invalidCfg.validate(); err == nil {
		t.Error("empty config should error")
	}
}
