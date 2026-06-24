package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/config"
	apperr "github.com/qw_pay/internal/errors"
	"github.com/qw_pay/internal/model"
)

type otpEntry struct {
	code      string
	createdAt time.Time
}

type Service struct {
	repo  *UserRepository
	mu    sync.RWMutex
	otp   map[string]*otpEntry
	ctx   context.Context
	cancel context.CancelFunc
}

func NewService(repo *UserRepository) *Service {
	s := &Service{repo: repo, otp: make(map[string]*otpEntry)}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	go s.cleanupOTP()
	return s
}

func (s *Service) cleanupOTP() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			for email, entry := range s.otp {
				if time.Since(entry.createdAt) > time.Duration(config.C.OTPTTLSeconds)*time.Second {
					delete(s.otp, email)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Service) Stop() {
	s.cancel()
}

func (s *Service) Register(ctx context.Context, email, phone, password string) (*model.User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to check email existence")
	}
	if exists {
		return nil, apperr.Conflict("email already registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to hash password")
	}

	user, err := s.repo.Create(ctx, email, phone, hash)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to create user")
	}
	return user, nil
}

func (s *Service) GenerateOTP() string {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			panic(fmt.Sprintf("failed to generate OTP digit: %v", err))
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code
}

func (s *Service) StoreOTP(email, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.otp[email] = &otpEntry{code: code, createdAt: time.Now()}
}

func (s *Service) VerifyOTP(email, code string) bool {
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
	if subtle.ConstantTimeCompare([]byte(entry.code), []byte(code)) != 1 {
		return false
	}
	delete(s.otp, email)
	return true
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*model.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}
	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return nil, apperr.Unauthorized("invalid credentials")
	}
	return user, nil
}

func (s *Service) VerifyUser(ctx context.Context, email string) error {
	if err := s.repo.SetVerified(ctx, email); err != nil {
		return apperr.Wrap(err, "failed to verify user")
	}
	return nil
}

func (s *Service) CreateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Duration(config.C.JWTExpireHours) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.C.JWTSecret))
	if err != nil {
		return "", apperr.Wrap(err, "failed to sign token")
	}
	return tokenStr, nil
}

func (s *Service) DecodeToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(config.C.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, apperr.Unauthorized("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, apperr.Unauthorized("invalid claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, apperr.Unauthorized("invalid subject")
	}
	parsedUUID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, apperr.Unauthorized("invalid user id")
	}
	return parsedUUID, nil
}
