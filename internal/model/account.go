package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountStatus string

const (
	StatusActive  AccountStatus = "ACTIVE"
	StatusBlocked AccountStatus = "BLOCKED"
)

type AccountCurrency string

const (
	CurrencyRUB AccountCurrency = "RUB"
	CurrencyUSD AccountCurrency = "USD"
	CurrencyEUR AccountCurrency = "EUR"
)

type Account struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Currency  AccountCurrency `json:"currency"`
	Balance   decimal.Decimal `json:"balance"`
	Version   int             `json:"version"`
	Status    AccountStatus   `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
