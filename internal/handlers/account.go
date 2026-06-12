package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/services"
)

// AccountHandler — хендлер для управления счетами.
type AccountHandler struct {
	accounts *services.AccountService
}

// NewAccountHandler создаёт новый экземпляр AccountHandler.
func NewAccountHandler(accounts *services.AccountService) *AccountHandler {
	return &AccountHandler{accounts: accounts}
}

type createAccountReq struct {
	Currency string `json:"currency" binding:"required,oneof=RUB USD EUR"`
}

// Create обрабатывает создание нового счёта.
func (h *AccountHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req createAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.accounts.Create(c.Request.Context(), userID, req.Currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}

// List обрабатывает получение списка счетов пользователя.
func (h *AccountHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	accounts, err := h.accounts.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

// Block обрабатывает блокировку счёта.
func (h *AccountHandler) Block(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}
	acc, err := h.accounts.GetByID(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	if acc.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your account"})
		return
	}
	if err := h.accounts.Block(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account blocked"})
}
