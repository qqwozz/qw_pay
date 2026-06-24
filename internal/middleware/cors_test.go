package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_Preflight(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Error("expected CORS header")
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Error("expected CORS header for allowed origin")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:8080")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("should not set CORS header for disallowed origin")
	}
}

func TestCORS_DefaultOrigin(t *testing.T) {
	os.Unsetenv("CORS_ALLOWED_ORIGINS")

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8080" {
		t.Error("expected default CORS origin")
	}
}
