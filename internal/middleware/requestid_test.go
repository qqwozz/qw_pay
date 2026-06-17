package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/contextkeys"
)

func TestRequestID_GeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id := c.GetString(string(contextkeys.KeyRequestID))
		if id == "" {
			t.Fatal("request_id should be in context")
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	r.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("X-Request-ID header should be set")
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("X-Request-ID should be a valid UUID: %v", err)
	}
}

func TestRequestID_PreservesExisting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id != "custom-id-123" {
			t.Fatalf("expected 'custom-id-123', got '%s'", id)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", "custom-id-123")
	r.ServeHTTP(w, req)
}

func TestRequestID_SetInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedID string

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		capturedID = c.GetString(string(contextkeys.KeyRequestID))
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	r.ServeHTTP(w, req)

	if capturedID == "" {
		t.Fatal("request_id should be set in context")
	}
	if _, err := uuid.Parse(capturedID); err != nil {
		t.Fatalf("request_id in context should be UUID: %v", err)
	}
}

func TestRequestID_DifferentRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ids := make(map[string]bool)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id := c.GetString(string(contextkeys.KeyRequestID))
		ids[id] = true
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		r.ServeHTTP(w, req)
	}

	if len(ids) != 10 {
		t.Errorf("expected 10 unique IDs, got %d", len(ids))
	}
}
