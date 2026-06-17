package antifraud

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qw_pay/internal/response"
)

type Handler struct {
	fraud *Client
}

func NewHandler(fraud *Client) *Handler {
	return &Handler{fraud: fraud}
}

type blockUserReq struct {
	UserID string `json:"user_id" binding:"required"`
}

type blockAccountReq struct {
	AccountID string `json:"account_id" binding:"required"`
}

func (h *Handler) BlockUser(c *gin.Context) {
	var req blockUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.fraud.BlockUser(c.Request.Context(), req.UserID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "User blocked in anti-fraud", "user_id": req.UserID})
}

func (h *Handler) BlockAccount(c *gin.Context) {
	var req blockAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.fraud.BlockAccount(c.Request.Context(), req.AccountID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Account blocked in anti-fraud", "account_id": req.AccountID})
}
