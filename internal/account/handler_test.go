package account

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
	config.C = &config.Config{
		MaxTransferAmount: decimal.NewFromInt(10_000_000),
		DailyLimit:        decimal.NewFromInt(50_000_000),
	}
}

func TestCreateReq_Binding(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		valid    bool
	}{
		{"USD", "USD", true},
		{"RUB", "RUB", true},
		{"EUR", "EUR", true},
		{"BTC", "BTC", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"currency": tt.currency})
			r := gin.New()
			r.POST("/test", func(c *gin.Context) {
				var req createReq
				err := c.ShouldBindJSON(&req)
				if tt.valid && err != nil {
					t.Errorf("expected valid, got error: %v", err)
				}
				if !tt.valid && err == nil {
					t.Error("expected error, got nil")
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
		})
	}
}
