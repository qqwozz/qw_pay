package audit

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"

	"github.com/qw_pay/internal/model"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(ctx context.Context, userID *uuid.UUID, action model.AuditAction, entityType string, entityID uuid.UUID, ip string) {
	entry := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  ip,
	}
	if err := s.repo.Log(ctx, entry); err != nil {
		log.Printf("[AUDIT] Failed to log: %v", err)
	}
}

func (s *Service) LogWithValue(ctx context.Context, userID *uuid.UUID, action model.AuditAction, entityType string, entityID uuid.UUID, ip string, oldVal, newVal interface{}) {
	entry := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  ip,
	}
	if oldVal != nil {
		b, _ := json.Marshal(oldVal)
		s := string(b)
		entry.OldValue = &s
	}
	if newVal != nil {
		b, _ := json.Marshal(newVal)
		s := string(b)
		entry.NewValue = &s
	}
	if err := s.repo.Log(ctx, entry); err != nil {
		log.Printf("[AUDIT] Failed to log: %v", err)
	}
}

func (s *Service) LogFraudVerdict(ctx context.Context, transactionID *uuid.UUID, verdict string, riskScore float64, engine string) {
	features := `{"auto": true}`
	check := &model.FraudCheck{
		TransactionID: transactionID,
		Verdict:       verdict,
		RiskScore:     riskScore,
		FeaturesJSON:  &features,
		Engine:        engine,
	}
	if err := s.repo.LogFraudCheck(ctx, check); err != nil {
		log.Printf("[AUDIT] Failed to log fraud check: %v", err)
	}
}

func (s *Service) List(ctx context.Context, userID *uuid.UUID, action string, page, pageSize int) ([]model.AuditLog, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, userID, action, pageSize, offset)
}
