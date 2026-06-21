package transaction

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByIDempotencyKey(ctx context.Context, key string) (*model.Transaction, error) {
	tx := &model.Transaction{}
	err := r.db.QueryRow(ctx,
		`SELECT id, idempotency_key, from_account_id, to_account_id, amount, currency,
		        source_currency, exchange_rate_used, status, created_at
		 FROM transactions WHERE idempotency_key=$1`, key,
	).Scan(&tx.ID, &tx.IdempotencyKey, &tx.FromAccountID, &tx.ToAccountID,
		&tx.Amount, &tx.Currency, &tx.SourceCurrency, &tx.ExchangeRate,
		&tx.Status, &tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get transaction by idempotency key: %w", err)
	}
	return tx, nil
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, key string, fromID, toID uuid.UUID, amount decimal.Decimal, sourceCurrency, targetCurrency string, exchangeRate *decimal.Decimal) (*model.Transaction, error) {
	t := &model.Transaction{}
	err := tx.QueryRow(ctx,
		`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'EXECUTED')
		 RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, status, created_at`,
		key, fromID, toID, amount, targetCurrency, sourceCurrency, exchangeRate,
	).Scan(&t.ID, &t.IdempotencyKey, &t.FromAccountID, &t.ToAccountID,
		&t.Amount, &t.Currency, &t.SourceCurrency, &t.ExchangeRate,
		&t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return t, nil
}

func (r *Repository) DebitAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error {
	tag, err := tx.Exec(ctx,
		`UPDATE accounts SET balance=balance-$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, accountID, version,
	)
	if err != nil {
		return fmt.Errorf("debit account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.ErrOptimisticLock
	}
	return nil
}

func (r *Repository) CreditAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount decimal.Decimal, version int) error {
	tag, err := tx.Exec(ctx,
		`UPDATE accounts SET balance=balance+$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, accountID, version,
	)
	if err != nil {
		return fmt.Errorf("credit account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.ErrOptimisticLock
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Transaction, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.idempotency_key, t.from_account_id, t.to_account_id, t.amount,
		        t.currency, t.source_currency, t.exchange_rate_used, t.status, t.created_at
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1
		 GROUP BY t.id
		 ORDER BY t.created_at DESC
		 OFFSET $2 LIMIT $3`,
		userID, offset, limit,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if scanErr := rows.Scan(&tx.ID, &tx.IdempotencyKey, &tx.FromAccountID, &tx.ToAccountID,
			&tx.Amount, &tx.Currency, &tx.SourceCurrency, &tx.ExchangeRate, &tx.Status, &tx.CreatedAt); scanErr != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", scanErr)
		}
		transactions = append(transactions, tx)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", rowsErr)
	}

	var total int
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT t.id)
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	return transactions, total, nil
}
