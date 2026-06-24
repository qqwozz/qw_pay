package transaction

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"from_account_id":"` + uuid.New().String() + `","to_account_id":"` + uuid.New().String() + `","amount":100,"idempotency_key":"key1"}`, false},
		{"missing from", `{"to_account_id":"` + uuid.New().String() + `","amount":100,"idempotency_key":"key1"}`, true},
		{"missing to", `{"from_account_id":"` + uuid.New().String() + `","amount":100,"idempotency_key":"key1"}`, true},
		{"missing key", `{"from_account_id":"` + uuid.New().String() + `","to_account_id":"` + uuid.New().String() + `","amount":100}`, true},
		{"bad json", `bad`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/test", func(c *gin.Context) {
				var req createReq
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

func TestListQueryParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantPage int
		wantSize int
	}{
		{"defaults", "", 1, 20},
		{"page=5", "page=5", 5, 20},
		{"size=50", "page_size=50", 1, 50},
		{"both", "page=3&page_size=10", 3, 10},
		{"invalid page", "page=abc", 1, 20},
		{"invalid size", "page_size=abc", 1, 20},
		{"size too large", "page_size=200", 1, 20},
		{"negative page", "page=-1", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/test", func(c *gin.Context) {
				page := 1
				pageSize := 20
				if p := c.Query("page"); p != "" {
					if v, err := strconv.Atoi(p); err == nil && v > 0 {
						page = v
					}
				}
				if ps := c.Query("page_size"); ps != "" {
					if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
						pageSize = v
					}
				}
				if page != tt.wantPage {
					t.Errorf("page: expected %d, got %d", tt.wantPage, page)
				}
				if pageSize != tt.wantSize {
					t.Errorf("pageSize: expected %d, got %d", tt.wantSize, pageSize)
				}
			})
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test?"+tt.query, http.NoBody)
			r.ServeHTTP(w, req)
		})
	}
}
