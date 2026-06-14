package currency

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListCurrencies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"currencies": h.svc.SupportedCurrencies(),
	})
}

func (h *Handler) GetRates(c *gin.Context) {
	rates, err := h.svc.GetAllRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

func (h *Handler) GetRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to query params required"})
		return
	}
	rate, err := h.svc.GetRate(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rate)
}

type convertReq struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
	From   string  `json:"from" binding:"required"`
	To     string  `json:"to" binding:"required"`
}

func (h *Handler) Convert(c *gin.Context) {
	var req convertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	converted, rate, err := h.svc.Convert(c.Request.Context(), req.Amount, req.From, req.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := gin.H{
		"original_amount":  req.Amount,
		"original_currency": req.From,
		"converted_amount": converted,
		"target_currency":  req.To,
	}
	if rate != nil {
		resp["exchange_rate"] = rate.Rate
	}
	c.JSON(http.StatusOK, resp)
}

type updateRateReq struct {
	From string  `json:"from" binding:"required"`
	To   string  `json:"to" binding:"required"`
	Rate float64 `json:"rate" binding:"required,gt=0"`
}

func (h *Handler) UpdateRate(c *gin.Context) {
	var req updateRateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateRate(c.Request.Context(), req.From, req.To, req.Rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rate updated"})
}
