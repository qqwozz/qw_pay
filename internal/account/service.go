package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/logger"
	"github.com/qw_pay/internal/model"
)

var WelcomeBonus = decimal.NewFromInt(100)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, currency string) (*model.Account, error) {
	acc, err := s.repo.Create(ctx, userID, currency, WelcomeBonus)
	if err != nil {
		return nil, err
	}
	logger.Info("account created", "id", acc.ID, "user", userID, "currency", currency, "bonus", WelcomeBonus)
	return acc, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Account, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Block(ctx context.Context, id uuid.UUID) error {
	return s.repo.Block(ctx, id)
}

func (s *Service) GetDailyTransferSum(ctx context.Context, accountID uuid.UUID) (decimal.Decimal, error) {
	return s.repo.GetDailyTransferSum(ctx, accountID)
}
