package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/response"
)

type Handler struct {
	svc       *Service
	refreshSvc *RefreshTokenService
}

func NewHandler(svc *Service, refreshSvc *RefreshTokenService) *Handler {
	return &Handler{svc: svc, refreshSvc: refreshSvc}
}

type registerReq struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type verifyOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp_code" binding:"required,len=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req.Email, req.Phone, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	otp := h.svc.GenerateOTP()
	h.svc.StoreOTP(req.Email, otp)
	slog.Info("OTP generated", "email", req.Email)
	response.Created(c, gin.H{
		"message": "Registration successful. Check logs for OTP code.",
		"user_id": user.ID,
	})
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if !h.svc.VerifyOTP(req.Email, req.OTP) {
		response.Error(c, http.StatusBadRequest, "Invalid or expired OTP")
		return
	}
	if err := h.svc.VerifyUser(c.Request.Context(), req.Email); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Account verified"})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.svc.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	if !user.IsVerified {
		response.Error(c, http.StatusForbidden, "Account not verified")
		return
	}
	pair, err := h.refreshSvc.CreateTokenPair(c.Request.Context(), user.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create token")
		return
	}
	response.OK(c, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"token_type":    pair.TokenType,
		"expires_in":    pair.ExpiresIn,
		"user_id":       user.ID,
	})
}
