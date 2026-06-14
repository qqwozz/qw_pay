package audit

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

func (r *Repository) Log(ctx context.Context, entry *model.AuditLog) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, entity_type, entity_id, old_value, new_value, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.UserID, entry.Action, entry.EntityType, entry.EntityID,
		entry.OldValue, entry.NewValue, entry.IPAddress)
	return err
}

func (r *Repository) List(ctx context.Context, userID *uuid.UUID, action string, limit, offset int) ([]model.AuditLog, int, error) {
	query := `SELECT id, user_id, action, entity_type, entity_id, old_value, new_value, ip_address, created_at
		      FROM audit_log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if userID != nil {
		query += ` AND user_id = $` + fmt.Sprintf("%d", argIdx)
		countQuery += ` AND user_id = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}

	if action != "" {
		query += ` AND action = $` + fmt.Sprintf("%d", argIdx)
		countQuery += ` AND action = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, action)
		argIdx++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID,
			&l.OldValue, &l.NewValue, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}

func (r *Repository) LogFraudCheck(ctx context.Context, check *model.FraudCheck) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO fraud_checks (transaction_id, verdict, risk_score, features_json, engine)
		 VALUES ($1, $2, $3, $4, $5)`,
		check.TransactionID, check.Verdict, check.RiskScore, check.FeaturesJSON, check.Engine)
	return err
}
