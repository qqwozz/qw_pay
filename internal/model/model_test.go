package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAccountStatus(t *testing.T) {
	if StatusActive != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", StatusActive)
	}
	if StatusBlocked != "BLOCKED" {
		t.Errorf("expected BLOCKED, got %s", StatusBlocked)
	}
}

func TestAccountCurrency(t *testing.T) {
	if CurrencyRUB != "RUB" {
		t.Errorf("expected RUB, got %s", CurrencyRUB)
	}
	if CurrencyUSD != "USD" {
		t.Errorf("expected USD, got %s", CurrencyUSD)
	}
	if CurrencyEUR != "EUR" {
		t.Errorf("expected EUR, got %s", CurrencyEUR)
	}
}

func TestAccountStruct(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	acc := Account{
		ID:        id,
		UserID:    userID,
		Currency:  CurrencyUSD,
		Balance:   100.50,
		Version:   1,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if acc.ID != id {
		t.Error("ID mismatch")
	}
	if acc.UserID != userID {
		t.Error("UserID mismatch")
	}
	if acc.Currency != CurrencyUSD {
		t.Error("Currency mismatch")
	}
	if acc.Balance != 100.50 {
		t.Error("Balance mismatch")
	}
	if acc.Version != 1 {
		t.Error("Version mismatch")
	}
	if acc.Status != StatusActive {
		t.Error("Status mismatch")
	}
}

func TestTransactionStatus(t *testing.T) {
	if TxStatusPending != "PENDING" {
		t.Errorf("expected PENDING, got %s", TxStatusPending)
	}
	if TxStatusExecuted != "EXECUTED" {
		t.Errorf("expected EXECUTED, got %s", TxStatusExecuted)
	}
	if TxStatusRejected != "REJECTED" {
		t.Errorf("expected REJECTED, got %s", TxStatusRejected)
	}
}

func TestTransactionStruct(t *testing.T) {
	id := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	now := time.Now()
	srcCurrency := "USD"
	exchangeRate := 90.91

	tx := Transaction{
		ID:             id,
		IdempotencyKey: "key-123",
		FromAccountID:  fromID,
		ToAccountID:    toID,
		Amount:         100.0,
		Currency:       "RUB",
		SourceCurrency: &srcCurrency,
		ExchangeRate:   &exchangeRate,
		Status:         TxStatusExecuted,
		CreatedAt:      now,
	}

	if tx.ID != id {
		t.Error("ID mismatch")
	}
	if tx.IdempotencyKey != "key-123" {
		t.Error("IdempotencyKey mismatch")
	}
	if tx.FromAccountID != fromID {
		t.Error("FromAccountID mismatch")
	}
	if tx.ToAccountID != toID {
		t.Error("ToAccountID mismatch")
	}
	if tx.Amount != 100.0 {
		t.Error("Amount mismatch")
	}
	if tx.Currency != "RUB" {
		t.Error("Currency mismatch")
	}
	if tx.SourceCurrency == nil || *tx.SourceCurrency != "USD" {
		t.Error("SourceCurrency mismatch")
	}
	if tx.ExchangeRate == nil || *tx.ExchangeRate != 90.91 {
		t.Error("ExchangeRate mismatch")
	}
	if tx.Status != TxStatusExecuted {
		t.Error("Status mismatch")
	}
}

func TestTransactionOptionalFields(t *testing.T) {
	tx := Transaction{
		ID:        uuid.New(),
		Amount:    50.0,
		Currency:  "RUB",
		Status:    TxStatusPending,
		CreatedAt: time.Now(),
	}

	if tx.SourceCurrency != nil {
		t.Error("SourceCurrency should be nil")
	}
	if tx.ExchangeRate != nil {
		t.Error("ExchangeRate should be nil")
	}
}

func TestUserRole(t *testing.T) {
	if RoleUser != "USER" {
		t.Errorf("expected USER, got %s", RoleUser)
	}
	if RoleAdmin != "ADMIN" {
		t.Errorf("expected ADMIN, got %s", RoleAdmin)
	}
}

func TestUserStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	user := User{
		ID:           id,
		Email:        "test@example.com",
		Phone:        "+79001234567",
		PasswordHash: "hashed-password",
		Role:         RoleUser,
		IsVerified:   true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if user.ID != id {
		t.Error("ID mismatch")
	}
	if user.Email != "test@example.com" {
		t.Error("Email mismatch")
	}
	if user.Phone != "+79001234567" {
		t.Error("Phone mismatch")
	}
	if user.PasswordHash != "hashed-password" {
		t.Error("PasswordHash mismatch")
	}
	if user.Role != RoleUser {
		t.Error("Role mismatch")
	}
	if !user.IsVerified {
		t.Error("IsVerified should be true")
	}
}
