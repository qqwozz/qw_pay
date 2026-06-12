package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/services"
)

// TransactionHandler — хендлер для управления переводами.
type TransactionHandler struct {
	txService *services.TransactionService
	accounts  *services.AccountService
}

// NewTransactionHandler создаёт новый экземпляр TransactionHandler.
func NewTransactionHandler(txService *services.TransactionService, accounts *services.AccountService) *TransactionHandler {
	return &TransactionHandler{txService: txService, accounts: accounts}
}

type createTransactionReq struct {
	FromAccountID  uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID    uuid.UUID `json:"to_account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}

// Create обрабатывает создание перевода между счетами.
func (h *TransactionHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req createTransactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Проверяем, что счёт-отправитель принадлежит пользователю
	from, err := h.accounts.GetByID(c.Request.Context(), req.FromAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		return
	}
	if from.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your account"})
		return
	}
	tx, err := h.txService.Create(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tx)
}

// List обрабатывает получение истории переводов пользователя.
func (h *TransactionHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page := 1
	pageSize := 20
	transactions, total, err := h.txService.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}
