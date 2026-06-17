package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestLimiter_AllowsRequests(t *testing.T) {
	l := New(10, 10)

	for i := 0; i < 5; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestLimiter_BlocksAfterBurst(t *testing.T) {
	l := New(1, 3)

	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if l.allow("1.2.3.4") {
		t.Error("request should be blocked after burst exhausted")
	}
}

func TestLimiter_DifferentIPs(t *testing.T) {
	l := New(1, 1)

	if !l.allow("1.1.1.1") {
		t.Error("first IP should be allowed")
	}
	if l.allow("1.1.1.1") {
		t.Error("first IP should be blocked")
	}
	if !l.allow("2.2.2.2") {
		t.Error("second IP should be allowed")
	}
}

func TestLimiter_Middleware(t *testing.T) {
	l := New(2, 2)
	r := gin.New()
	r.Use(l.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "1.2.3.4:1234"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestLimiter_MiddlewareDifferentIPs(t *testing.T) {
	l := New(1, 1)
	r := gin.New()
	r.Use(l.Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", http.NoBody)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatal("first IP should be allowed")
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", http.NoBody)
	req2.RemoteAddr = "10.0.0.2:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatal("different IP should be allowed")
	}
}

func TestNewLimiter(t *testing.T) {
	l := New(100, 200)
	if l == nil {
		t.Fatal("limiter should not be nil")
	}
	if l.rate != 100 {
		t.Errorf("expected rate 100, got %d", l.rate)
	}
	if l.burst != 200 {
		t.Errorf("expected burst 200, got %d", l.burst)
	}
}
