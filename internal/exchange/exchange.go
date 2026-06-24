package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/logger"
)

const (
	frankfurterURL = "https://api.frankfurter.dev/v1"
	cacheTTL       = 1 * time.Hour
)

type cachedRates struct {
	rates     map[string]decimal.Decimal
	expiresAt time.Time
}

type Provider struct {
	mu      sync.RWMutex
	cache   map[string]*cachedRates
	client  *http.Client
	baseURL string
}

func NewProvider() *Provider {
	return &Provider{
		cache:   make(map[string]*cachedRates),
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: frankfurterURL,
	}
}

type frankfurterResponse struct {
	Base  string            `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func (p *Provider) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	p.mu.RLock()
	if cached, ok := p.cache[from]; ok && time.Now().Before(cached.expiresAt) {
		if rate, ok := cached.rates[to]; ok {
			p.mu.RUnlock()
			return rate, nil
		}
	}
	p.mu.RUnlock()

	rates, err := p.fetchRates(ctx, from)
	if err != nil {
		return decimal.Zero, fmt.Errorf("fetch rates for %s: %w", from, err)
	}

	p.mu.Lock()
	p.cache[from] = &cachedRates{
		rates:     rates,
		expiresAt: time.Now().Add(cacheTTL),
	}
	p.mu.Unlock()

	rate, ok := rates[to]
	if !ok {
		return decimal.Zero, fmt.Errorf("no exchange rate for %s to %s", from, to)
	}
	return rate, nil
}

func (p *Provider) fetchRates(ctx context.Context, base string) (map[string]decimal.Decimal, error) {
	url := fmt.Sprintf("%s/latest?base=%s", p.baseURL, base)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	rates := make(map[string]decimal.Decimal, len(result.Rates))
	for code, rate := range result.Rates {
		d, err := decimal.NewFromString(fmt.Sprintf("%f", rate))
		if err != nil {
			logger.Warn("failed to parse exchange rate", "currency", code, "rate", rate)
			continue
		}
		rates[code] = d
	}

	logger.Info("exchange rates fetched", "base", base, "count", len(rates))
	return rates, nil
}

func (p *Provider) GetRateWithFallback(ctx context.Context, from, to string, fallback map[string]decimal.Decimal) (decimal.Decimal, error) {
	rate, err := p.GetRate(ctx, from, to)
	if err != nil {
		logger.Warn("exchange rate fetch failed, using fallback", "from", from, "to", to, "error", err)
		key := from + "_" + to
		if r, ok := fallback[key]; ok {
			return r, nil
		}
		return decimal.Zero, fmt.Errorf("no exchange rate for %s to %s", from, to)
	}
	return rate, nil
}
