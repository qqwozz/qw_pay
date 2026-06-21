package contextkeys

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	KeyUserID    contextKey = "user_id"
	KeyRequestID contextKey = "request_id"
)

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(string(KeyUserID))
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
