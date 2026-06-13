package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, currency string, balance float64) (*model.Account, error) {
	acc := &model.Account{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO accounts (user_id, currency, balance, version, status)
		 VALUES ($1, $2, $3, 1, 'ACTIVE')
		 RETURNING id, user_id, currency, balance, version, status, created_at, updated_at`,
		userID, currency, balance,
	).Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return acc, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Account, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, currency, balance, version, status, created_at, updated_at
		 FROM accounts WHERE user_id=$1 ORDER BY created_at`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var acc model.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *Repository) Block(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE accounts SET status='BLOCKED', updated_at=NOW() WHERE id=$1 AND status='ACTIVE'`, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account already blocked")
	}
	return nil
}

func (r *Repository) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var sum float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE from_account_id=$1 AND created_at >= DATE_TRUNC('day', NOW())
		 AND status='EXECUTED'`, accountID,
	).Scan(&sum)
	return sum, err
}
