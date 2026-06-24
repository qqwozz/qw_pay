package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRegisterReq_Binding(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"email":"test@example.com","phone":"+123","password":"secret123"}`, false},
		{"missing email", `{"phone":"+123","password":"secret123"}`, true},
		{"missing phone", `{"email":"test@example.com","password":"secret123"}`, true},
		{"short password", `{"email":"test@example.com","phone":"+123","password":"ab"}`, true},
		{"bad email", `{"email":"not-email","phone":"+123","password":"secret123"}`, true},
		{"bad json", `bad`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/test", func(c *gin.Context) {
				var req registerReq
				err := c.ShouldBindJSON(&req)
				if tt.wantErr && err == nil {
					t.Error("expected error, got nil")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("expected valid, got: %v", err)
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
		})
	}
}

func TestVerifyOTPReq_Binding(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"email":"test@example.com","otp_code":"123456"}`, false},
		{"missing email", `{"otp_code":"123456"}`, true},
		{"wrong len", `{"email":"test@example.com","otp_code":"123"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/test", func(c *gin.Context) {
				var req verifyOTPReq
				err := c.ShouldBindJSON(&req)
				if tt.wantErr && err == nil {
					t.Error("expected error, got nil")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("expected valid, got: %v", err)
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
		})
	}
}

func TestLoginReq_Binding(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"email":"test@example.com","password":"secret123"}`, false},
		{"missing email", `{"password":"secret123"}`, true},
		{"missing password", `{"email":"test@example.com"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/test", func(c *gin.Context) {
				var req loginReq
				err := c.ShouldBindJSON(&req)
				if tt.wantErr && err == nil {
					t.Error("expected error, got nil")
				}
				if !tt.wantErr && err != nil {
					t.Errorf("expected valid, got: %v", err)
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte(tt.body)))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
		})
	}
}

func TestRefreshReq_Binding(t *testing.T) {
	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		var req refreshReq
		err := c.ShouldBindJSON(&req)
		if err == nil {
			t.Error("expected error for empty body")
		}
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
}
