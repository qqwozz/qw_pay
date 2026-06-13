package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/model"
)

var ErrOptimisticLock = errors.New("optimistic lock conflict")

var exchangeRates = map[string]float64{
	"RUB_USD": 0.011,
	"USD_RUB": 90.91,
	"RUB_EUR": 0.010,
	"EUR_RUB": 100.0,
	"USD_EUR": 0.92,
	"EUR_USD": 1.09,
}

type accountReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (float64, error)
}

type Service struct {
	db   *pgxpool.Pool
	repo *Repository
	acc  accountReader
}

func NewService(db *pgxpool.Pool, repo *Repository, acc accountReader) *Service {
	return &Service{db: db, repo: repo, acc: acc}
}

func (s *Service) Create(ctx context.Context, fromID, toID uuid.UUID, amount float64, idempotencyKey string) (*model.Transaction, error) {
	existing, err := s.repo.GetByIDempotencyKey(ctx, idempotencyKey)
	if err == nil {
		return existing, nil
	}

	from, err := s.acc.GetByID(ctx, fromID)
	if err != nil {
		return nil, fmt.Errorf("source account not found")
	}
	to, err := s.acc.GetByID(ctx, toID)
	if err != nil {
		return nil, fmt.Errorf("target account not found")
	}

	if from.Status != model.StatusActive || to.Status != model.StatusActive {
		return nil, fmt.Errorf("account is blocked")
	}
	if amount > config.C.MaxTransferAmount {
		return nil, fmt.Errorf("amount exceeds max transfer limit")
	}
	if fromID == toID {
		return nil, fmt.Errorf("cannot transfer to the same account")
	}

	dailySum, err := s.acc.GetDailyTransferSum(ctx, fromID)
	if err != nil {
		return nil, err
	}
	if dailySum+amount > config.C.DailyLimit {
		return nil, fmt.Errorf("daily transfer limit exceeded")
	}

	var exchangeRate *float64
	sourceCurrency := string(from.Currency)
	targetCurrency := string(to.Currency)

	if sourceCurrency != targetCurrency {
		key := sourceCurrency + "_" + targetCurrency
		rate, ok := exchangeRates[key]
		if !ok {
			return nil, fmt.Errorf("no exchange rate for %s to %s", sourceCurrency, targetCurrency)
		}
		exchangeRate = &rate
	}

	dbTx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback(ctx)

	transaction, err := s.repo.Create(ctx, dbTx, idempotencyKey, fromID, toID, amount, sourceCurrency, targetCurrency, exchangeRate)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DebitAccount(ctx, dbTx, fromID, amount, from.Version); err != nil {
		return nil, err
	}

	if err := s.repo.CreditAccount(ctx, dbTx, toID, amount, to.Version); err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, err
	}

	log.Printf("Transaction executed: id=%s from=%s to=%s amount=%.2f %s → %s",
		transaction.ID, fromID, toID, amount, sourceCurrency, targetCurrency)
	return transaction, nil
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Transaction, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, offset, pageSize)
}
