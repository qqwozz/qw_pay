package currency

import (
	"context"
	"fmt"
	"log"
)

var supportedCurrencies = []string{"RUB", "USD", "EUR"}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	if from == to {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         1.0,
			Source:       "identity",
		}, nil
	}
	return s.repo.GetRate(ctx, from, to)
}

func (s *Service) Convert(ctx context.Context, amount float64, from, to string) (float64, *ExchangeRate, error) {
	if from == to {
		return amount, nil, nil
	}
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, nil, fmt.Errorf("no exchange rate for %s to %s", from, to)
	}
	converted := amount * rate.Rate
	log.Printf("Currency conversion: %.2f %s -> %.2f %s (rate=%.6f)",
		amount, from, converted, to, rate.Rate)
	return converted, rate, nil
}

func (s *Service) GetAllRates(ctx context.Context) ([]ExchangeRate, error) {
	return s.repo.GetAllRates(ctx)
}

func (s *Service) UpdateRate(ctx context.Context, from, to string, rate float64) error {
	if from == to {
		return fmt.Errorf("cannot set rate for same currency")
	}
	err := s.repo.UpsertRate(ctx, from, to, rate, "api")
	if err == nil {
		log.Printf("Exchange rate updated: %s -> %s = %.6f", from, to, rate)
	}
	return err
}

func (s *Service) SupportedCurrencies() []string {
	return supportedCurrencies
}
