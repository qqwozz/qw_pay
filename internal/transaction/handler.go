package transaction

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/audit"
)

type Handler struct {
	svc    *Service
	acc    *account.Service
	fraud  *antifraud.Client
	audit  *audit.Service
}

func NewHandler(svc *Service, acc *account.Service, fraud *antifraud.Client, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, acc: acc, fraud: fraud, audit: auditSvc}
}

type createReq struct {
	FromAccountID  uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID    uuid.UUID `json:"to_account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	SourceCurrency string    `json:"source_currency"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	from, err := h.acc.GetByID(c.Request.Context(), req.FromAccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source account not found"})
		return
	}
	if from.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your account"})
		return
	}

	if h.fraud != nil {
		verdict, err := h.fraud.Check(
			c.Request.Context(),
			req.FromAccountID.String(),
			req.ToAccountID.String(),
			req.Amount,
			string(from.Currency),
			userID.String(),
		)
		if err != nil {
			log.Printf("[ANTIFRAUD] Check failed: %v — proceeding without anti-fraud", err)
		} else if !verdict.Approved {
			log.Printf("[ANTIFRAUD] Transaction BLOCKED: id=%s reason=%s risk=%d",
				verdict.ID, verdict.Reason, verdict.RiskScore)
			if h.audit != nil {
				h.audit.Log(c.Request.Context(), &userID, "TRANSFER_BLOCKED_BY_FRAUD", "transaction", uuid.Nil, c.ClientIP())
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Transaction blocked by anti-fraud system",
				"reason":     verdict.Reason,
				"risk":       verdict.RiskScore,
				"verdict_id": verdict.ID,
			})
			return
		} else {
			log.Printf("[ANTIFRAUD] Transaction APPROVED: id=%s risk=%d engine=%s",
				verdict.ID, verdict.RiskScore, verdict.Engine)
		}
	}

	tx, err := h.svc.Create(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.SourceCurrency, req.IdempotencyKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.audit != nil {
		h.audit.Log(c.Request.Context(), &userID, "TRANSFER_COMPLETED", "transaction", tx.ID, c.ClientIP())
	}
	c.JSON(http.StatusOK, tx)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	transactions, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
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
