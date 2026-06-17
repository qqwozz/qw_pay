package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/contextkeys"
	"github.com/qw_pay/internal/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createReq struct {
	Currency string `json:"currency" binding:"required,oneof=RUB USD EUR"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	acc, err := h.svc.Create(c.Request.Context(), userID, req.Currency)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, acc)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)
	accounts, err := h.svc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, accounts)
}

func (h *Handler) Block(c *gin.Context) {
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid account ID")
		return
	}
	acc, err := h.svc.GetByID(c.Request.Context(), accountID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Account not found")
		return
	}
	if acc.UserID != userID {
		response.Error(c, http.StatusForbidden, "Not your account")
		return
	}
	if err := h.svc.Block(c.Request.Context(), accountID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Account blocked"})
}
