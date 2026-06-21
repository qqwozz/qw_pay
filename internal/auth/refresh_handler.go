package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/contextkeys"
	"github.com/qw_pay/internal/response"
)

type RefreshHandler struct {
	svc *RefreshTokenService
}

func NewRefreshHandler(svc *RefreshTokenService) *RefreshHandler {
	return &RefreshHandler{svc: svc}
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *RefreshHandler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	pair, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.OK(c, pair)
}

func (h *RefreshHandler) Logout(c *gin.Context) {
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)

	if err := h.svc.Logout(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "logged out"})
}
