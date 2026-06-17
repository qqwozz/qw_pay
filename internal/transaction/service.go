package transaction

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	apperr "github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/model"
)

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

type txRepository interface {
	GetByIDempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	Create(ctx context.Context, tx pgx.Tx, key string, fromID, toID uuid.UUID, amount float64, sourceCurrency, targetCurrency string, exchangeRate *float64) (*model.Transaction, error)
	DebitAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount float64, version int) error
	CreditAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount float64, version int) error
	ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Transaction, int, error)
}

type Service struct {
	db   *pgxpool.Pool
	repo txRepository
	acc  accountReader
}

func NewService(db *pgxpool.Pool, repo txRepository, acc accountReader) *Service {
	return &Service{db: db, repo: repo, acc: acc}
}

func (s *Service) Create(ctx context.Context, fromID, toID uuid.UUID, amount float64, idempotencyKey string) (*model.Transaction, error) {
	existing, err := s.repo.GetByIDempotencyKey(ctx, idempotencyKey)
	if err == nil {
		return existing, nil
	}

	from, err := s.acc.GetByID(ctx, fromID)
	if err != nil {
		return nil, apperr.NotFound("source account not found")
	}
	to, err := s.acc.GetByID(ctx, toID)
	if err != nil {
		return nil, apperr.NotFound("target account not found")
	}

	if from.Status != model.StatusActive || to.Status != model.StatusActive {
		return nil, apperr.Forbidden("account is blocked")
	}
	if amount <= 0 {
		return nil, apperr.BadRequest("amount must be positive")
	}
	if amount > config.C.MaxTransferAmount {
		return nil, apperr.BadRequest(fmt.Sprintf("amount exceeds max transfer limit of %.2f", config.C.MaxTransferAmount))
	}
	if fromID == toID {
		return nil, apperr.BadRequest("cannot transfer to the same account")
	}

	dailySum, err := s.acc.GetDailyTransferSum(ctx, fromID)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to get daily transfer sum")
	}
	if dailySum+amount > config.C.DailyLimit {
		return nil, apperr.BadRequest("daily transfer limit exceeded")
	}

	sourceCurrency := string(from.Currency)
	targetCurrency := string(to.Currency)
	var exchangeRate *float64

	if sourceCurrency != targetCurrency {
		key := sourceCurrency + "_" + targetCurrency
		rate, ok := exchangeRates[key]
		if !ok {
			return nil, apperr.BadRequest(fmt.Sprintf("no exchange rate for %s to %s", sourceCurrency, targetCurrency))
		}
		exchangeRate = &rate
	}

	dbTx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to begin transaction")
	}
	defer func() { _ = dbTx.Rollback(ctx) }()

	transaction, err := s.repo.Create(ctx, dbTx, idempotencyKey, fromID, toID, amount, sourceCurrency, targetCurrency, exchangeRate)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to create transaction record")
	}

	if err := s.repo.DebitAccount(ctx, dbTx, fromID, amount, from.Version); err != nil {
		return nil, apperr.Wrap(err, "failed to debit source account")
	}

	if err := s.repo.CreditAccount(ctx, dbTx, toID, amount, to.Version); err != nil {
		return nil, apperr.Wrap(err, "failed to credit target account")
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, apperr.Wrap(err, "failed to commit transaction")
	}

	slog.Info("transaction executed",
		"id", transaction.ID,
		"from", fromID,
		"to", toID,
		"amount", amount,
		"from_currency", sourceCurrency,
		"to_currency", targetCurrency,
	)
	return transaction, nil
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, offset, pageSize)
}
