package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/currency"
	"github.com/qw_pay/internal/model"
)

var ErrOptimisticLock = errors.New("optimistic lock conflict")

const maxRetries = 3

type accountReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (float64, error)
}

type Service struct {
	db   *pgxpool.Pool
	repo *Repository
	acc  accountReader
	cur  *currency.Service
}

func NewService(db *pgxpool.Pool, repo *Repository, acc accountReader, cur *currency.Service) *Service {
	return &Service{db: db, repo: repo, acc: acc, cur: cur}
}

func (s *Service) Create(ctx context.Context, fromID, toID uuid.UUID, amount float64, sourceCurrency, idempotencyKey string) (*model.Transaction, error) {
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

	if err := s.validateTransfer(from, to, fromID, toID, amount); err != nil {
		return nil, err
	}

	if sourceCurrency == "" {
		sourceCurrency = string(from.Currency)
	}
	if sourceCurrency != string(from.Currency) {
		return nil, fmt.Errorf("source_currency %s does not match account currency %s", sourceCurrency, from.Currency)
	}

	targetCurrency := string(to.Currency)
	convertedAmount, exchangeRate, err := s.convertIfNeeded(ctx, amount, sourceCurrency, targetCurrency)
	if err != nil {
		return nil, err
	}

	if err := s.checkLimits(ctx, fromID, amount); err != nil {
		return nil, err
	}

	return s.executeTransfer(ctx, fromID, toID, from, to, amount, sourceCurrency, targetCurrency, exchangeRate, convertedAmount, idempotencyKey)
}

func (s *Service) validateTransfer(from, to *model.Account, fromID, toID uuid.UUID, amount float64) error {
	if from.Status != model.StatusActive || to.Status != model.StatusActive {
		return fmt.Errorf("account is blocked")
	}
	if fromID == toID {
		return fmt.Errorf("cannot transfer to the same account")
	}
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if from.Balance < amount {
		return fmt.Errorf("insufficient funds: balance=%.2f, amount=%.2f", from.Balance, amount)
	}
	return nil
}

func (s *Service) convertIfNeeded(ctx context.Context, amount float64, sourceCurrency, targetCurrency string) (float64, *float64, error) {
	if sourceCurrency == targetCurrency {
		return amount, nil, nil
	}
	converted, rateInfo, err := s.cur.Convert(ctx, amount, sourceCurrency, targetCurrency)
	if err != nil {
		return 0, nil, err
	}
	var rate *float64
	if rateInfo != nil {
		rate = &rateInfo.Rate
	}
	return converted, rate, nil
}

func (s *Service) checkLimits(ctx context.Context, fromID uuid.UUID, amount float64) error {
	if amount > config.C.MaxTransferAmount {
		return fmt.Errorf("amount exceeds max transfer limit")
	}
	dailySum, err := s.acc.GetDailyTransferSum(ctx, fromID)
	if err != nil {
		return err
	}
	if dailySum+amount > config.C.DailyLimit {
		return fmt.Errorf("daily transfer limit exceeded")
	}
	return nil
}

func (s *Service) executeTransfer(ctx context.Context, fromID, toID uuid.UUID, from, to *model.Account, amount float64, sourceCurrency, targetCurrency string, exchangeRate *float64, convertedAmount float64, idempotencyKey string) (*model.Transaction, error) {
	for i := 0; i < maxRetries; i++ {
		dbTx, err := s.db.Begin(ctx)
		if err != nil {
			return nil, err
		}

		transaction, err := s.repo.Create(ctx, dbTx, idempotencyKey, fromID, toID, amount, sourceCurrency, targetCurrency, exchangeRate, convertedAmount)
		if err != nil {
			dbTx.Rollback(ctx)
			return nil, err
		}

		if err := s.repo.DebitAccount(ctx, dbTx, fromID, amount, from.Version); err != nil {
			dbTx.Rollback(ctx)
			if errors.Is(err, ErrOptimisticLock) {
				from, to, err = s.refreshAccounts(ctx, fromID, toID)
				if err != nil {
					return nil, err
				}
				log.Printf("Optimistic lock on debit, retry %d/%d", i+1, maxRetries)
				continue
			}
			return nil, err
		}

		if err := s.repo.CreditAccount(ctx, dbTx, toID, convertedAmount, to.Version); err != nil {
			dbTx.Rollback(ctx)
			if errors.Is(err, ErrOptimisticLock) {
				from, to, err = s.refreshAccounts(ctx, fromID, toID)
				if err != nil {
					return nil, err
				}
				log.Printf("Optimistic lock on credit, retry %d/%d", i+1, maxRetries)
				continue
			}
			return nil, err
		}

		if err := dbTx.Commit(ctx); err != nil {
			return nil, err
		}

		log.Printf("Transaction executed: id=%s from=%s to=%s %.2f %s -> %.2f %s (rate=%v)",
			transaction.ID, fromID, toID, amount, sourceCurrency, convertedAmount, targetCurrency, exchangeRate)
		return transaction, nil
	}
	return nil, ErrOptimisticLock
}

func (s *Service) refreshAccounts(ctx context.Context, fromID, toID uuid.UUID) (*model.Account, *model.Account, error) {
	from, err := s.acc.GetByID(ctx, fromID)
	if err != nil {
		return nil, nil, err
	}
	to, err := s.acc.GetByID(ctx, toID)
	if err != nil {
		return nil, nil, err
	}
	return from, to, nil
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Transaction, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, offset, pageSize)
}
