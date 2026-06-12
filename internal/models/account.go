package models

import (
	"time"

	"github.com/google/uuid"
)

// AccountStatus — статус счёта.
type AccountStatus string

const (
	StatusActive  AccountStatus = "ACTIVE"  // Счёт активен
	StatusBlocked AccountStatus = "BLOCKED" // Счёт заблокирован
)

// AccountCurrency — поддерживаемые валюты.
type AccountCurrency string

const (
	CurrencyRUB AccountCurrency = "RUB" // Российский рубль
	CurrencyUSD AccountCurrency = "USD" // Доллар США
	CurrencyEUR AccountCurrency = "EUR" // Евро
)

// Account — банковский счёт пользователя.
type Account struct {
	ID        uuid.UUID       `json:"id"`         // Уникальный идентификатор счёта
	UserID    uuid.UUID       `json:"user_id"`    // Владелец счёта
	Currency  AccountCurrency `json:"currency"`   // Валюта счёта
	Balance   float64         `json:"balance"`    // Текущий баланс
	Version   int             `json:"version"`    // Версия для optimistic locking
	Status    AccountStatus   `json:"status"`     // Статус счёта
	CreatedAt time.Time       `json:"created_at"` // Дата создания
	UpdatedAt time.Time       `json:"updated_at"` // Дата последнего обновления
}
