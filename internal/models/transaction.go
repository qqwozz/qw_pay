package models

import (
	"time"

	"github.com/google/uuid"
)

// TransactionStatus — статус транзакции.
type TransactionStatus string

const (
	TxStatusPending  TransactionStatus = "PENDING"  // Ожидает обработки
	TxStatusExecuted TransactionStatus = "EXECUTED" // Успешно выполнена
	TxStatusRejected TransactionStatus = "REJECTED" // Отклонена
)

// Transaction — платёжная транзакция (перевод между счетами).
type Transaction struct {
	ID             uuid.UUID         `json:"id"`               // Уникальный идентификатор
	IdempotencyKey string            `json:"idempotency_key"` // Ключ идемпотентности (уникальный)
	FromAccountID  uuid.UUID         `json:"from_account_id"`  // Счёт-отправитель
	ToAccountID    uuid.UUID         `json:"to_account_id"`    // Счёт-получатель
	Amount         float64           `json:"amount"`           // Сумма перевода
	Currency       string            `json:"currency"`         // Валюта зачисления
	SourceCurrency *string           `json:"source_currency,omitempty"`   // Валюта списания (при конвертации)
	ExchangeRate   *float64          `json:"exchange_rate_used,omitempty"` // Курс конвертации
	Status         TransactionStatus `json:"status"`           // Статус транзакции
	CreatedAt      time.Time         `json:"created_at"`       // Дата создания
}
