package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/audit"
)

type Handler struct {
	svc   *Service
	audit *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

type createReq struct {
	Currency string `json:"currency" binding:"required,oneof=RUB USD EUR"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.svc.Create(c.Request.Context(), userID, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if h.audit != nil {
		h.audit.Log(c.Request.Context(), &userID, "ACCOUNT_CREATED", "account", acc.ID, c.ClientIP())
	}
	c.JSON(http.StatusOK, acc)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	accounts, err := h.svc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) Block(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	acc, err := h.svc.GetByID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	if acc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your account"})
		return
	}
	if err := h.svc.Block(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.audit != nil {
		h.audit.Log(c.Request.Context(), &userID, "ACCOUNT_BLOCKED", "account", accountID, c.ClientIP())
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account blocked"})
}

func (h *Handler) AdminBlock(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	acc, err := h.svc.GetByID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	if err := h.svc.Block(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID := c.MustGet("user_id").(uuid.UUID)
	if h.audit != nil {
		h.audit.LogWithValue(c.Request.Context(), &adminID, "ACCOUNT_BLOCKED", "account", accountID, c.ClientIP(),
			map[string]string{"status": string(acc.Status)}, map[string]string{"status": "BLOCKED"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account blocked by admin"})
}
