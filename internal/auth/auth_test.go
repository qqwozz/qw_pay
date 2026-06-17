package auth

import (
	"strings"
	"testing"

	"github.com/qw_pay/internal/config"
)

func init() {
	config.C = &config.Config{
		JWTSecret:      "test-secret-for-auth",
		JWTExpireHours: 24,
		OTPTTLSeconds:  300,
	}
}

func TestHashPassword(t *testing.T) {
	t.Run("hashes password successfully", func(t *testing.T) {
		hash, err := HashPassword("secret123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hash == "" {
			t.Error("hash should not be empty")
		}
		if hash == "secret123" {
			t.Error("hash should not equal plaintext")
		}
		if !strings.HasPrefix(hash, "$2a$") {
			t.Error("hash should start with $2a$")
		}
	})

	t.Run("different hashes for same password", func(t *testing.T) {
		hash1, _ := HashPassword("test")
		hash2, _ := HashPassword("test")
		if hash1 == hash2 {
			t.Error("bcrypt should produce different hashes")
		}
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("correct password", func(t *testing.T) {
		hash, _ := HashPassword("correct")
		err := CheckPassword(hash, "correct")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, _ := HashPassword("correct")
		err := CheckPassword(hash, "wrong")
		if err == nil {
			t.Error("expected error for wrong password")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		hash, _ := HashPassword("password")
		err := CheckPassword(hash, "")
		if err == nil {
			t.Error("expected error for empty password")
		}
	})
}

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	if svc == nil {
		t.Fatal("service should not be nil")
	}
	if svc.otp == nil {
		t.Error("otp map should be initialized")
	}
}

func TestGenerateOTP(t *testing.T) {
	svc := NewService(nil)

	t.Run("generates 6-digit code", func(t *testing.T) {
		otp := svc.GenerateOTP()
		if len(otp) != 6 {
			t.Errorf("expected 6 digits, got %d", len(otp))
		}
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("expected digit, got %c", c)
			}
		}
	})

	t.Run("different calls produce different codes", func(t *testing.T) {
		otp1 := svc.GenerateOTP()
		otp2 := svc.GenerateOTP()
		if otp1 == otp2 {
			t.Error("OTP codes should differ")
		}
	})
}

func TestStoreAndVerifyOTP(t *testing.T) {
	svc := NewService(nil)

	t.Run("store and verify", func(t *testing.T) {
		svc.StoreOTP("test@example.com", "123456")
		if !svc.VerifyOTP("test@example.com", "123456") {
			t.Error("OTP should verify")
		}
	})

	t.Run("OTP deleted after verify", func(t *testing.T) {
		svc.StoreOTP("test2@example.com", "654321")
		svc.VerifyOTP("test2@example.com", "654321")
		if svc.VerifyOTP("test2@example.com", "654321") {
			t.Error("OTP should be deleted after first verify")
		}
	})

	t.Run("wrong OTP fails", func(t *testing.T) {
		svc.StoreOTP("test3@example.com", "111111")
		if svc.VerifyOTP("test3@example.com", "999999") {
			t.Error("wrong OTP should fail")
		}
	})

	t.Run("non-existent email fails", func(t *testing.T) {
		if svc.VerifyOTP("nonexistent@example.com", "000000") {
			t.Error("non-existent email should fail")
		}
	})
}

func TestCreateToken(t *testing.T) {
	svc := NewService(nil)
	t.Run("creates valid token", func(t *testing.T) {
		token, err := svc.CreateToken([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Error("token should not be empty")
		}
	})
}

func TestDecodeToken(t *testing.T) {
	svc := NewService(nil)

	t.Run("roundtrip valid token", func(t *testing.T) {
		token, err := svc.CreateToken([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		if err != nil {
			t.Fatalf("create token error: %v", err)
		}
		userID, err := svc.DecodeToken(token)
		if err != nil {
			t.Fatalf("decode token error: %v", err)
		}
		expected := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		if userID != expected {
			t.Errorf("expected %v, got %v", expected, userID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := svc.DecodeToken("invalid.token.here")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := svc.DecodeToken("")
		if err == nil {
			t.Error("expected error for empty token")
		}
	})
}
