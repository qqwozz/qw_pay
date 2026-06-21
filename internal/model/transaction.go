package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TxStatusPending  TransactionStatus = "PENDING"
	TxStatusExecuted TransactionStatus = "EXECUTED"
	TxStatusRejected TransactionStatus = "REJECTED"
)

type Transaction struct {
	ID             uuid.UUID          `json:"id"`
	IdempotencyKey string             `json:"idempotency_key"`
	FromAccountID  uuid.UUID          `json:"from_account_id"`
	ToAccountID    uuid.UUID          `json:"to_account_id"`
	Amount         decimal.Decimal    `json:"amount"`
	Currency       string             `json:"currency"`
	SourceCurrency *string            `json:"source_currency,omitempty"`
	ExchangeRate   *decimal.Decimal   `json:"exchange_rate_used,omitempty"`
	Status         TransactionStatus  `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
}
