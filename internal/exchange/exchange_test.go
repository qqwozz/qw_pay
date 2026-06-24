package exchange

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/logger"
)

func init() {
	logger.Setup(&bytes.Buffer{})
	slog.SetDefault(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))
}

func TestGetRate_SameCurrency(t *testing.T) {
	p := NewProvider()
	rate, err := p.GetRate(context.Background(), "USD", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("expected 1, got %s", rate)
	}
}

func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := frankfurterResponse{
			Base: "USD",
			Rates: map[string]float64{
				"EUR": 0.85,
				"GBP": 0.73,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	rate, err := p.GetRate(context.Background(), "USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(decimal.RequireFromString("0.85")) {
		t.Errorf("expected 0.85, got %s", rate)
	}
}

func TestGetRate_MissingCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := frankfurterResponse{
			Base: "USD",
			Rates: map[string]float64{
				"EUR": 0.85,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	_, err := p.GetRate(context.Background(), "USD", "JPY")
	if err == nil {
		t.Error("expected error for missing currency")
	}
}

func TestGetRate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	_, err := p.GetRate(context.Background(), "USD", "EUR")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestGetRate_Cached(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := frankfurterResponse{
			Base: "USD",
			Rates: map[string]float64{
				"EUR": 0.85,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	p.GetRate(context.Background(), "USD", "EUR")
	p.GetRate(context.Background(), "USD", "EUR")
	p.GetRate(context.Background(), "USD", "EUR")

	if callCount != 1 {
		t.Errorf("expected 1 HTTP call (cached), got %d", callCount)
	}
}

func TestGetRateWithFallback_Fallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	fallback := map[string]decimal.Decimal{
		"USD_EUR": decimal.RequireFromString("0.85"),
	}

	rate, err := p.GetRateWithFallback(context.Background(), "USD", "EUR", fallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(decimal.RequireFromString("0.85")) {
		t.Errorf("expected fallback 0.85, got %s", rate)
	}
}

func TestGetRateWithFallback_NoFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := &Provider{
		cache:   make(map[string]*cachedRates),
		client:  server.Client(),
		baseURL: server.URL,
	}

	fallback := map[string]decimal.Decimal{}

	_, err := p.GetRateWithFallback(context.Background(), "USD", "EUR", fallback)
	if err == nil {
		t.Error("expected error when no fallback available")
	}
}
