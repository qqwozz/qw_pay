package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/config"
	apperr "github.com/qw_pay/internal/errors"
)

type RefreshTokenService struct {
	repo *RefreshTokenRepository
}

func NewRefreshTokenService(repo *RefreshTokenRepository) *RefreshTokenService {
	return &RefreshTokenService{repo: repo}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *RefreshTokenService) CreateTokenPair(ctx context.Context, userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := createAccessToken(userID)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to create access token")
	}

	_, refreshToken, err := s.repo.Create(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(err, "failed to create refresh token")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    config.C.JWTExpireHours * 3600,
	}, nil
}

func (s *RefreshTokenService) Refresh(ctx context.Context, rawRefreshToken string) (*TokenPair, error) {
	rt, err := s.repo.GetByToken(ctx, rawRefreshToken)
	if err != nil {
		return nil, apperr.Unauthorized("invalid refresh token")
	}

	if rt.Revoked {
		return nil, apperr.Unauthorized("refresh token revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, apperr.Unauthorized("refresh token expired")
	}

	if err := s.repo.Revoke(ctx, rt.ID); err != nil {
		return nil, apperr.Wrap(err, "failed to revoke old refresh token")
	}

	return s.CreateTokenPair(ctx, rt.UserID)
}

func (s *RefreshTokenService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllForUser(ctx, userID)
}

func createAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Duration(config.C.JWTExpireHours) * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.C.JWTSecret))
}
