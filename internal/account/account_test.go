package account

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/model"
)

func TestServiceConstants(t *testing.T) {
	if !WelcomeBonus.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected WelcomeBonus 100, got %s", WelcomeBonus)
	}
}

func TestNewService(t *testing.T) {
	repo := &Repository{db: nil}
	svc := NewService(repo)
	if svc == nil {
		t.Fatal("service should not be nil")
	}
	if svc.repo != repo {
		t.Error("repo should be set")
	}
}

type mockAccountReader struct {
	accounts map[uuid.UUID]*model.Account
	dailySum decimal.Decimal
}

func (m *mockAccountReader) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	if acc, ok := m.accounts[id]; ok {
		return acc, nil
	}
	return nil, errors.ErrNotFound
}

func (m *mockAccountReader) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	return m.dailySum, nil
}

func TestRepository_NewRepository(t *testing.T) {
	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("repository should not be nil")
	}
}

func TestAccountJSON(t *testing.T) {
	acc := model.Account{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Currency:  model.CurrencyUSD,
		Balance:   decimal.RequireFromString("1000.00"),
		Version:   5,
		Status:    model.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if acc.Currency != model.CurrencyUSD {
		t.Error("currency should be USD")
	}
	if !acc.Balance.Equal(decimal.RequireFromString("1000.00")) {
		t.Error("balance should be 1000.00")
	}
	if acc.Version != 5 {
		t.Error("version should be 5")
	}
	if acc.Status != model.StatusActive {
		t.Error("status should be ACTIVE")
	}
}

func TestMockAccountReader(t *testing.T) {
	userID := uuid.New()
	accID := uuid.New()

	mock := &mockAccountReader{
		accounts: map[uuid.UUID]*model.Account{
			accID: {
				ID:       accID,
				UserID:   userID,
				Currency: model.CurrencyUSD,
				Balance:  decimal.RequireFromString("500.0"),
				Version:  1,
				Status:   model.StatusActive,
			},
		},
		dailySum: decimal.RequireFromString("100.0"),
	}

	t.Run("GetByID found", func(t *testing.T) {
		acc, err := mock.GetByID(context.Background(), accID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if acc.ID != accID {
			t.Error("ID mismatch")
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := mock.GetByID(context.Background(), uuid.New())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("GetDailyTransferSum", func(t *testing.T) {
		sum, err := mock.GetDailyTransferSum(context.Background(), accID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !sum.Equal(decimal.RequireFromString("100.0")) {
			t.Errorf("expected 100.0, got %s", sum)
		}
	})
}

func TestAccountStatusConstants(t *testing.T) {
	if model.StatusActive != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", model.StatusActive)
	}
	if model.StatusBlocked != "BLOCKED" {
		t.Errorf("expected BLOCKED, got %s", model.StatusBlocked)
	}
}

func TestAccountCurrencyConstants(t *testing.T) {
	currencies := []model.AccountCurrency{
		model.CurrencyRUB,
		model.CurrencyUSD,
		model.CurrencyEUR,
	}
	for _, c := range currencies {
		if c == "" {
			t.Error("currency should not be empty")
		}
	}
}
