package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/contextkeys"
)

func init() {
	config.C = &config.Config{
		JWTSecret:      "test-secret-key-for-middleware",
		JWTExpireHours: 24,
	}
}

func createTestToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTSecret))
}

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing Authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", http.NoBody)

		AuthRequired()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid Authorization format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", http.NoBody)
		c.Request.Header.Set("Authorization", "Basic abc123")

		AuthRequired()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid Bearer token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", http.NoBody)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		AuthRequired()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("valid token sets user_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", http.NoBody)

		userID := uuid.New()
		token, err := createTestToken(userID.String())
		if err != nil {
			t.Fatalf("failed to create test token: %v", err)
		}
		c.Request.Header.Set("Authorization", "Bearer "+token)

		AuthRequired()(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		storedUserID, exists := c.Get(string(contextkeys.KeyUserID))
		if !exists {
			t.Error("user_id should be set in context")
		}
		if storedUserID == nil {
			t.Error("user_id should not be nil")
		}
	})

	t.Run("empty Bearer token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", http.NoBody)
		c.Request.Header.Set("Authorization", "Bearer ")

		AuthRequired()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})
}
