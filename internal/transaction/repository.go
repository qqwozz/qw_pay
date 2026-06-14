package transaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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
		        source_currency, exchange_rate_used, converted_amount, status, created_at
		 FROM transactions WHERE idempotency_key=$1`, key,
	).Scan(&tx.ID, &tx.IdempotencyKey, &tx.FromAccountID, &tx.ToAccountID,
		&tx.Amount, &tx.Currency, &tx.SourceCurrency, &tx.ExchangeRate,
		&tx.ConvertedAmount, &tx.Status, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, key string, fromID, toID uuid.UUID, amount float64, sourceCurrency, targetCurrency string, exchangeRate *float64, convertedAmount float64) (*model.Transaction, error) {
	t := &model.Transaction{}
	err := tx.QueryRow(ctx,
		`INSERT INTO transactions (idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, converted_amount, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'EXECUTED')
		 RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, source_currency, exchange_rate_used, converted_amount, status, created_at`,
		key, fromID, toID, amount, targetCurrency, sourceCurrency, exchangeRate, convertedAmount,
	).Scan(&t.ID, &t.IdempotencyKey, &t.FromAccountID, &t.ToAccountID,
		&t.Amount, &t.Currency, &t.SourceCurrency, &t.ExchangeRate,
		&t.ConvertedAmount, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *Repository) DebitAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount float64, version int) error {
	tag, err := tx.Exec(ctx,
		`UPDATE accounts SET balance=balance-$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, accountID, version,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOptimisticLock
	}
	return nil
}

func (r *Repository) CreditAccount(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, amount float64, version int) error {
	tag, err := tx.Exec(ctx,
		`UPDATE accounts SET balance=balance+$1, version=version+1, updated_at=NOW()
		 WHERE id=$2 AND version=$3 AND status='ACTIVE'`,
		amount, accountID, version,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOptimisticLock
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.Transaction, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.idempotency_key, t.from_account_id, t.to_account_id, t.amount,
		        t.currency, t.source_currency, t.exchange_rate_used, t.converted_amount, t.status, t.created_at
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1
		 GROUP BY t.id
		 ORDER BY t.created_at DESC
		 OFFSET $2 LIMIT $3`,
		userID, offset, limit,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(&tx.ID, &tx.IdempotencyKey, &tx.FromAccountID, &tx.ToAccountID,
			&tx.Amount, &tx.Currency, &tx.SourceCurrency, &tx.ExchangeRate,
			&tx.ConvertedAmount, &tx.Status, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}

	var total int
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(DISTINCT t.id)
		 FROM transactions t
		 JOIN accounts a ON a.id = t.from_account_id OR a.id = t.to_account_id
		 WHERE a.user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
