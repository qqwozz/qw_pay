package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/config"
	apperr "github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/logger"
	"github.com/qw_pay/internal/model"
)

var fallbackRates = map[string]decimal.Decimal{
	"RUB_USD": decimal.RequireFromString("0.011"),
	"USD_RUB": decimal.RequireFromString("90.91"),
	"RUB_EUR": decimal.RequireFromString("0.010"),
	"EUR_RUB": decimal.RequireFromString("100.0"),
	"USD_EUR": decimal.RequireFromString("0.92"),
	"EUR_USD": decimal.RequireFromString("1.09"),
}

type accountReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error)
}

type txRepository interface {
	GetByIDempotencyKey(ctx context.Context, key string) (*model.Transaction, error)
	Create(ctx context.Context, tx pgx.Tx, key string, fromID, toID uuid.UUID, amount decimal.Decimal, sourceCurrency, targetCurrency string, exchangeRate *decimal.Decimal) (*model.Transaction, error)
	DebitAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error
	CreditAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error
	ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Transaction, int, error)
}

type exchangeReader interface {
	GetRateWithFallback(ctx context.Context, from, to string, fallback map[string]decimal.Decimal) (decimal.Decimal, error)
}

type Service struct {
	db       *pgxpool.Pool
	repo     txRepository
	acc      accountReader
	exchange exchangeReader
}

func NewService(db *pgxpool.Pool, repo txRepository, acc accountReader, exchange exchangeReader) *Service {
	return &Service{db: db, repo: repo, acc: acc, exchange: exchange}
}

func (s *Service) Create(ctx context.Context, fromID, toID uuid.UUID, amount decimal.Decimal, idempotencyKey string) (*model.Transaction, error) {
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
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.BadRequest("amount must be positive")
	}
	if amount.GreaterThan(config.C.MaxTransferAmount) {
		return nil, apperr.BadRequest(fmt.Sprintf("amount exceeds max transfer limit of %s", config.C.MaxTransferAmount))
	}
	if fromID == toID {
		return nil, apperr.BadRequest("cannot transfer to the same account")
	}
	if from.Balance.LessThan(amount) {
		return nil, apperr.BadRequest("insufficient funds")
	}

	dailySum, err := s.acc.GetDailyTransferSum(ctx, fromID)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to get daily transfer sum")
	}
	if dailySum.Add(amount).GreaterThan(config.C.DailyLimit) {
		return nil, apperr.BadRequest("daily transfer limit exceeded")
	}

	sourceCurrency := string(from.Currency)
	targetCurrency := string(to.Currency)
	var exchangeRate *decimal.Decimal

	creditAmount := amount
	if sourceCurrency != targetCurrency {
		rate, rateErr := s.exchange.GetRateWithFallback(ctx, sourceCurrency, targetCurrency, fallbackRates)
		if rateErr != nil {
			return nil, apperr.BadRequest(fmt.Sprintf("no exchange rate for %s to %s", sourceCurrency, targetCurrency))
		}
		exchangeRate = &rate
		creditAmount = amount.Mul(rate)
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

	if err := s.repo.CreditAccount(ctx, dbTx, toID, creditAmount, to.Version); err != nil {
		return nil, apperr.Wrap(err, "failed to credit target account")
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, apperr.Wrap(err, "failed to commit transaction")
	}

	logger.Info("transaction executed",
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
