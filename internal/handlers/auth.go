package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/services"
)

// AuthHandler — хендлер для эндпоинтов аутентификации.
type AuthHandler struct {
	auth *services.AuthService
}

// NewAuthHandler создаёт новый экземпляр AuthHandler.
func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
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

// Register обрабатывает регистрацию нового пользователя.
// После успешной регистрации генерируется OTP-код (выводится в лог).
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.auth.Register(c.Request.Context(), req.Email, req.Phone, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Генерируем и сохраняем OTP, выводим в лог (для MVP)
	otp := h.auth.GenerateOTP()
	h.auth.StoreOTP(req.Email, otp)
	log.Printf("OTP for %s: %s", req.Email, otp)
	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful. Check logs for OTP code.",
		"user_id": user.ID,
	})
}

// VerifyOTP обрабатывает подтверждение аккаунта через OTP-код.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.auth.VerifyOTP(req.Email, req.OTP) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP"})
		return
	}
	if err := h.auth.VerifyUser(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account verified"})
}

// Login обрабатывает вход в систему и выдаёт JWT-токен.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.auth.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account not verified"})
		return
	}
	token, err := h.auth.CreateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "bearer",
		"user_id":      user.ID,
	})
}
