package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, currency string, balance decimal.Decimal) (*model.Account, error) {
	acc := &model.Account{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO accounts (user_id, currency, balance, version, status)
		 VALUES ($1, $2, $3, 1, 'ACTIVE')
		 RETURNING id, user_id, currency, balance, version, status, created_at, updated_at`,
		userID, currency, balance,
	).Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return acc, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	acc := &model.Account{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, currency, balance, version, status, created_at, updated_at
		 FROM accounts WHERE id=$1`, id,
	).Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get account by id: %w", err)
	}
	return acc, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Account, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, currency, balance, version, status, created_at, updated_at
		 FROM accounts WHERE user_id=$1 ORDER BY created_at`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var acc model.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

func (r *Repository) Block(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE accounts SET status='BLOCKED', updated_at=NOW() WHERE id=$1 AND status='ACTIVE'`, id,
	)
	if err != nil {
		return fmt.Errorf("block account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account already blocked")
	}
	return nil
}

func (r *Repository) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	var sum decimal.Decimal
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE from_account_id=$1 AND created_at >= DATE_TRUNC('day', NOW())
		 AND status='EXECUTED'`, accountID,
	).Scan(&sum)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get daily transfer sum: %w", err)
	}
	return sum, nil
}
