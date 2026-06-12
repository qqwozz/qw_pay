package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/models"
)

// AccountService отвечает за управление банковскими счетами.
type AccountService struct {
	db *pgxpool.Pool
}

// NewAccountService создаёт новый экземпляр AccountService.
func NewAccountService(db *pgxpool.Pool) *AccountService {
	return &AccountService{db: db}
}

// Create создаёт новый счёт для пользователя в указанной валюте.
// При создании начисляется приветственный бонус 100 единиц валюты.
func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, currency string) (*models.Account, error) {
	const welcomeBonus = 100.0

	acc := &models.Account{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO accounts (user_id, currency, balance, version, status)
		 VALUES ($1, $2, $3, 1, 'ACTIVE')
		 RETURNING id, user_id, currency, balance, version, status, created_at, updated_at`,
		userID, currency, welcomeBonus,
	).Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, err
	}

	log.Printf("Account created: id=%s user=%s currency=%s bonus=%.2f", acc.ID, userID, currency, welcomeBonus)
	return acc, nil
}

// GetByID возвращает счёт по его ID.
func (s *AccountService) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	acc := &models.Account{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, currency, balance, version, status, created_at, updated_at
		 FROM accounts WHERE id=$1`, id,
	).Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

// ListByUser возвращает все счета пользователя, отсортированные по дате создания.
func (s *AccountService) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, currency, balance, version, status, created_at, updated_at
		 FROM accounts WHERE user_id=$1 ORDER BY created_at`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Currency, &acc.Balance, &acc.Version, &acc.Status, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

// Block блокирует счёт (только если он активен).
func (s *AccountService) Block(ctx context.Context, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx,
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

// GetDailyTransferSum возвращает сумму исходящих переводов за текущий день.
func (s *AccountService) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var sum float64
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE from_account_id=$1 AND created_at >= DATE_TRUNC('day', NOW())
		 AND status='EXECUTED'`, accountID,
	).Scan(&sum)
	return sum, err
}
