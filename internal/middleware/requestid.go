package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/qw_pay/internal/contextkeys"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(string(contextkeys.KeyRequestID), id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
