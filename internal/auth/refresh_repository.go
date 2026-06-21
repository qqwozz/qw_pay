package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/model"
)

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID) (*model.RefreshToken, string, error) {
	rawToken, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}

	hash := hashToken(rawToken)
	expiresAt := time.Now().Add(time.Duration(config.C.RefreshExpireDays) * 24 * time.Hour)

	rt := &model.RefreshToken{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, created_at, revoked`,
		userID, hash, expiresAt,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		return nil, "", fmt.Errorf("create refresh token: %w", err)
	}

	return rt, rawToken, nil
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, rawToken string) (*model.RefreshToken, error) {
	hash := hashToken(rawToken)
	rt := &model.RefreshToken{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked
		 FROM refresh_tokens WHERE token_hash=$1`, hash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET revoked=true WHERE id=$1", id)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_tokens SET revoked=true WHERE user_id=$1", userID)
	return err
}

func (r *RefreshTokenRepository) Cleanup(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		"DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked=true")
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
