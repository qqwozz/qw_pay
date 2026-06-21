package transaction

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/account"
	"github.com/qw_pay/internal/antifraud"
	"github.com/qw_pay/internal/contextkeys"
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
	FromAccountID  uuid.UUID `json:"from_account_id" binding:"required"`
	ToAccountID    uuid.UUID `json:"to_account_id" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)
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
			req.Amount,
			string(from.Currency),
			userID.String(),
		)
		switch {
		case err != nil:
			slog.Warn("antifraud check failed, proceeding", "error", err)
		case !verdict.Approved:
			slog.Warn("transaction blocked",
				"verdict_id", verdict.ID,
				"reason", verdict.Reason,
				"risk", verdict.RiskScore,
			)
			response.Error(c, http.StatusForbidden, "Transaction blocked by anti-fraud system")
			return
		default:
			slog.Info("transaction approved",
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
	userID := c.MustGet(string(contextkeys.KeyUserID)).(uuid.UUID)
	page := 1
	pageSize := 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}
	transactions, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, transactions, page, pageSize, total)
}
