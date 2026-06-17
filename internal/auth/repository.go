package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/qw_pay/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email existence: %w", err)
	}
	return exists, nil
}

func (r *UserRepository) Create(ctx context.Context, email, phone, passwordHash string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, phone, password_hash, role, is_verified)
		 VALUES ($1, $2, $3, 'USER', false)
		 RETURNING id, email, phone, role, is_verified, created_at, updated_at`,
		email, phone, passwordHash,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.Role, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, phone, password_hash, role, is_verified, created_at, updated_at
		 FROM users WHERE email=$1`, email,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Role, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) SetVerified(ctx context.Context, email string) error {
	tag, err := r.db.Exec(ctx, "UPDATE users SET is_verified=true, updated_at=NOW() WHERE email=$1", email)
	if err != nil {
		return fmt.Errorf("set verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
