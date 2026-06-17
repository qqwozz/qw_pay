package antifraud

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	fraud *Client
}

func NewHandler(fraud *Client) *Handler {
	return &Handler{fraud: fraud}
}

func (h *Handler) BlockUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.fraud.BlockUser(c.Request.Context(), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User blocked in anti-fraud", "user_id": req.UserID})
}

func (h *Handler) BlockAccount(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.fraud.BlockAccount(c.Request.Context(), req.AccountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account blocked in anti-fraud", "account_id": req.AccountID})
}
