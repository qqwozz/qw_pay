package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/models"
)

// otpEntry хранит OTP-код и время его создания.
type otpEntry struct {
	code      string
	createdAt time.Time
}

// AuthService отвечает за регистрацию, аутентификацию и управление сессиями.
type AuthService struct {
	db  *pgxpool.Pool
	mu  sync.RWMutex
	otp map[string]*otpEntry // In-memory хранилище OTP (для MVP)
}

// NewAuthService создаёт новый экземпляр AuthService.
func NewAuthService(db *pgxpool.Pool) *AuthService {
	return &AuthService{db: db, otp: make(map[string]*otpEntry)}
}

// Register регистрирует нового пользователя в системе.
// Возвращает ошибку, если email уже занят.
func (s *AuthService) Register(ctx context.Context, email, phone, password string) (*models.User, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	err = s.db.QueryRow(ctx,
		`INSERT INTO users (email, phone, password_hash, role, is_verified)
		 VALUES ($1, $2, $3, 'USER', false)
		 RETURNING id, email, phone, role, is_verified, created_at, updated_at`,
		email, phone, string(hash),
	).Scan(&user.ID, &user.Email, &user.Phone, &user.Role, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GenerateOTP генерирует 6-значный код подтверждения.
func (s *AuthService) GenerateOTP() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code
}

// StoreOTP сохраняет OTP-код в памяти с привязкой к email.
func (s *AuthService) StoreOTP(email, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.otp[email] = &otpEntry{code: code, createdAt: time.Now()}
}

// VerifyOTP проверяет OTP-код для указанного email.
// Код действителен в течение OTPTTLSeconds.
func (s *AuthService) VerifyOTP(email, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.otp[email]
	if !ok {
		return false
	}
	if time.Since(entry.createdAt) > time.Duration(config.C.OTPTTLSeconds)*time.Second {
		delete(s.otp, email)
		return false
	}
	if entry.code != code {
		return false
	}
	delete(s.otp, email)
	return true
}

// Authenticate проверяет email/пароль и возвращает пользователя.
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(ctx,
		`SELECT id, email, phone, password_hash, role, is_verified, created_at, updated_at
		 FROM users WHERE email=$1`, email,
	).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Role, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

// VerifyUser подтверждает аккаунт пользователя (после ввода OTP).
func (s *AuthService) VerifyUser(ctx context.Context, email string) error {
	_, err := s.db.Exec(ctx, "UPDATE users SET is_verified=true, updated_at=NOW() WHERE email=$1", email)
	return err
}

// CreateToken создаёт JWT-токен для авторизации.
func (s *AuthService) CreateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Duration(config.C.JWTExpireHours) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTSecret))
}

// DecodeToken проверяет JWT-токен и возвращает user_id.
func (s *AuthService) DecodeToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.C.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid subject")
	}
	return uuid.Parse(sub)
}
