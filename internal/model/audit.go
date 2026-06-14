package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditTransferCreated       AuditAction = "TRANSFER_CREATED"
	AuditTransferCompleted     AuditAction = "TRANSFER_COMPLETED"
	AuditTransferFailed        AuditAction = "TRANSFER_FAILED"
	AuditTransferBlockedFraud  AuditAction = "TRANSFER_BLOCKED_BY_FRAUD"
	AuditAccountCreated        AuditAction = "ACCOUNT_CREATED"
	AuditAccountBlocked        AuditAction = "ACCOUNT_BLOCKED"
	AuditAccountUnblocked      AuditAction = "ACCOUNT_UNBLOCKED"
	AuditUserRegistered        AuditAction = "USER_REGISTERED"
	AuditUserVerified          AuditAction = "USER_VERIFIED"
)

type AuditLog struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Action     AuditAction `json:"action"`
	EntityType string     `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	OldValue   *string    `json:"old_value,omitempty"`
	NewValue   *string    `json:"new_value,omitempty"`
	IPAddress  string     `json:"ip_address"`
	CreatedAt  time.Time  `json:"created_at"`
}

type FraudCheck struct {
	ID            uuid.UUID  `json:"id"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	Verdict       string     `json:"verdict"`
	RiskScore     float64    `json:"risk_score"`
	FeaturesJSON  *string    `json:"features_json,omitempty"`
	Engine        string     `json:"engine"`
	CheckedAt     time.Time  `json:"checked_at"`
}
