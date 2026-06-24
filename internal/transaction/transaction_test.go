package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/model"
)

func init() {
	config.C = &config.Config{
		MaxTransferAmount: decimal.NewFromInt(10_000_000),
		DailyLimit:        decimal.NewFromInt(50_000_000),
	}
}

func TestExchangeRates(t *testing.T) {
	expected := map[string]decimal.Decimal{
		"RUB_USD": decimal.RequireFromString("0.011"),
		"USD_RUB": decimal.RequireFromString("90.91"),
		"RUB_EUR": decimal.RequireFromString("0.010"),
		"EUR_RUB": decimal.RequireFromString("100.0"),
		"USD_EUR": decimal.RequireFromString("0.92"),
		"EUR_USD": decimal.RequireFromString("1.09"),
	}

	for key, expectedRate := range expected {
		if rate, ok := fallbackRates[key]; !ok {
			t.Errorf("missing exchange rate for %s", key)
		} else if !rate.Equal(expectedRate) {
			t.Errorf("exchange rate %s: expected %s, got %s", key, expectedRate, rate)
		}
	}
}

func TestExchangeRateKeys(t *testing.T) {
	rates := []string{"RUB_USD", "USD_RUB", "RUB_EUR", "EUR_RUB", "USD_EUR", "EUR_USD"}
	if len(fallbackRates) != len(rates) {
		t.Errorf("expected %d rates, got %d", len(rates), len(fallbackRates))
	}
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("repository should not be nil")
	}
}

func TestNewService(t *testing.T) {
	repo := &Repository{db: nil}
	svc := NewService(nil, repo, nil, nil)
	if svc == nil {
		t.Fatal("service should not be nil")
	}
}

type mockAccount struct {
	account *model.Account
	err     error
}

type mockAccountService struct {
	accounts map[uuid.UUID]*mockAccount
	dailySum decimal.Decimal
}

func (m *mockAccountService) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	if acc, ok := m.accounts[id]; ok {
		return acc.account, acc.err
	}
	return nil, errors.ErrNotFound
}

func (m *mockAccountService) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	return m.dailySum, nil
}

type mockTxRepo struct {
	txByIDempotencyKey func(key string) (*model.Transaction, error)
}

func (m *mockTxRepo) GetByIDempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	if m.txByIDempotencyKey != nil {
		return m.txByIDempotencyKey(key)
	}
	return nil, errors.ErrNotFound
}

func (m *mockTxRepo) Create(ctx context.Context, tx pgx.Tx, key string, fromID, toID uuid.UUID, amount decimal.Decimal, sourceCurrency, targetCurrency string, exchangeRate *decimal.Decimal) (*model.Transaction, error) {
	return nil, nil
}

func (m *mockTxRepo) DebitAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error {
	return nil
}

func (m *mockTxRepo) CreditAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error {
	return nil
}

func (m *mockTxRepo) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Transaction, int, error) {
	return nil, 0, nil
}

type mockExchange struct {
	rate decimal.Decimal
	err  error
}

func (m *mockExchange) GetRateWithFallback(ctx context.Context, from, to string, fallback map[string]decimal.Decimal) (decimal.Decimal, error) {
	if m.err != nil {
		return decimal.Zero, m.err
	}
	return m.rate, nil
}

func TestService_Create_SourceNotFound(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key4")
	if err == nil {
		t.Error("expected error for unknown source account")
	}
}

func TestService_Create_TargetNotFound(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key5")
	if err == nil {
		t.Error("expected error for unknown target account")
	}
}

func TestService_Create_SameAccount(t *testing.T) {
	fromID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, fromID, decimal.NewFromInt(100), "key3")
	if err == nil {
		t.Error("expected error for same account transfer")
	}
}

func TestService_Create_BlockedAccount(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusBlocked, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key-blocked")
	if err == nil {
		t.Error("expected error for blocked source account")
	}
}

func TestService_Create_BlockedTargetAccount(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusBlocked, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key-blocked2")
	if err == nil {
		t.Error("expected error for blocked target account")
	}
}

func TestService_Create_MaxAmount(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(20_000_000), "key-max")
	if err == nil {
		t.Error("expected error for amount exceeding max transfer limit")
	}
}

func TestService_Create_DailyLimit(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.RequireFromString("49999999"),
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(10), "key-daily")
	if err == nil {
		t.Error("expected error for daily limit exceeded")
	}
}

func TestService_Create_NoExchangeRate(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: "GBP"}},
		},
		dailySum: decimal.Zero,
	}

	mockExchange := &mockExchange{err: errors.ErrNotFound}
	svc := NewService(nil, &mockTxRepo{}, mockAcc, mockExchange)
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key-rate")
	if err == nil {
		t.Error("expected error for missing exchange rate")
	}
}

func TestService_Create_ZeroAmount(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.Zero, "key-zero")
	if err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestService_Create_NegativeAmount(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(-10), "key-neg")
	if err == nil {
		t.Error("expected error for negative amount")
	}
}

func TestService_Create_InsufficientFunds(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()

	mockAcc := &mockAccountService{
		accounts: map[uuid.UUID]*mockAccount{
			fromID: {account: &model.Account{ID: fromID, Status: model.StatusActive, Currency: model.CurrencyRUB, Balance: decimal.NewFromInt(50)}},
			toID:   {account: &model.Account{ID: toID, Status: model.StatusActive, Currency: model.CurrencyRUB}},
		},
		dailySum: decimal.Zero,
	}

	svc := NewService(nil, &mockTxRepo{}, mockAcc, &mockExchange{rate: decimal.NewFromInt(1)})
	_, err := svc.Create(context.Background(), fromID, toID, decimal.NewFromInt(100), "key-funds")
	if err == nil {
		t.Error("expected error for insufficient funds")
	}
}

func TestService_ListByUser_PaginationDefaults(t *testing.T) {
	svc := NewService(nil, &mockTxRepo{}, nil, nil)

	t.Run("page 0 defaults to 1", func(t *testing.T) {
		_, _, err := svc.ListByUser(context.Background(), uuid.New(), 0, 20)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("pageSize 0 defaults to 20", func(t *testing.T) {
		_, _, err := svc.ListByUser(context.Background(), uuid.New(), 1, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("pageSize > 100 defaults to 20", func(t *testing.T) {
		_, _, err := svc.ListByUser(context.Background(), uuid.New(), 1, 200)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("negative page defaults to 1", func(t *testing.T) {
		_, _, err := svc.ListByUser(context.Background(), uuid.New(), -1, 10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestTransactionModel(t *testing.T) {
	tx := model.Transaction{
		ID:             uuid.New(),
		IdempotencyKey: "test-key",
		FromAccountID:  uuid.New(),
		ToAccountID:    uuid.New(),
		Amount:         decimal.NewFromInt(100),
		Currency:       "RUB",
		Status:         model.TxStatusExecuted,
	}

	if !tx.Amount.Equal(decimal.NewFromInt(100)) {
		t.Error("amount should be 100")
	}
	if tx.Currency != "RUB" {
		t.Error("currency should be RUB")
	}
	if tx.Status != model.TxStatusExecuted {
		t.Error("status should be EXECUTED")
	}
}

func TestTransactionStatusConstants(t *testing.T) {
	if model.TxStatusPending != "PENDING" {
		t.Errorf("expected PENDING, got %s", model.TxStatusPending)
	}
	if model.TxStatusExecuted != "EXECUTED" {
		t.Errorf("expected EXECUTED, got %s", model.TxStatusExecuted)
	}
	if model.TxStatusRejected != "REJECTED" {
		t.Errorf("expected REJECTED, got %s", model.TxStatusRejected)
	}
}

func TestErrOptimisticLock(t *testing.T) {
	if errors.ErrOptimisticLock == nil {
		t.Error("ErrOptimisticLock should not be nil")
	}
	if errors.ErrOptimisticLock.Error() != "optimistic lock conflict" {
		t.Errorf("unexpected error message: %s", errors.ErrOptimisticLock.Error())
	}
}
