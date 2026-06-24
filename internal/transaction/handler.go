package transaction

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/contextkeys"
	"github.com/qw_pay/internal/logger"
	"github.com/qw_pay/internal/response"
)

type Handler struct {
	svc   *Service
	acc   *account.Service
	fraud *antifraud.Client
}

func NewHandler(svc *Service, acc *account.Service, fraud *antifraud.Client) *Handler {
	return &Handler{svc: svc, acc: acc, fraud: fraud}
}

type createReq struct {
	FromAccountID  uuid.UUID       `json:"from_account_id" binding:"required"`
	ToAccountID    uuid.UUID       `json:"to_account_id" binding:"required"`
	Amount         decimal.Decimal `json:"amount" binding:"required"`
	IdempotencyKey string          `json:"idempotency_key" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := contextkeys.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	from, err := h.acc.GetByID(c.Request.Context(), req.FromAccountID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Source account not found")
		return
	}
	if from.UserID != userID {
		response.Error(c, http.StatusForbidden, "Not your account")
		return
	}

	if h.fraud != nil {
		var verdict *antifraud.Verdict
		verdict, err = h.fraud.Check(
			c.Request.Context(),
			req.FromAccountID.String(),
			req.ToAccountID.String(),
			req.Amount.InexactFloat64(),
			string(from.Currency),
			userID.String(),
		)
		switch {
		case err != nil:
			logger.Warn("antifraud check failed, proceeding", "error", err)
		case !verdict.Approved:
			logger.Warn("transaction blocked",
				"verdict_id", verdict.ID,
				"reason", verdict.Reason,
				"risk", verdict.RiskScore,
			)
			response.Error(c, http.StatusForbidden, "Transaction blocked by anti-fraud system")
			return
		default:
			logger.Info("transaction approved",
				"verdict_id", verdict.ID,
				"risk", verdict.RiskScore,
				"engine", verdict.Engine,
			)
		}
	}

	tx, err := h.svc.Create(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount, req.IdempotencyKey)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(c, tx)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := contextkeys.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	transactions, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, transactions, page, pageSize, total)
}
