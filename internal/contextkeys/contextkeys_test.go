package contextkeys

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetUserID_Found(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", http.NoBody)

	expectedID := uuid.New()
	c.Set(string(KeyUserID), expectedID)

	id, ok := GetUserID(c)
	if !ok {
		t.Error("expected ok=true")
	}
	if id != expectedID {
		t.Errorf("expected %v, got %v", expectedID, id)
	}
}

func TestGetUserID_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", http.NoBody)

	id, ok := GetUserID(c)
	if ok {
		t.Error("expected ok=false")
	}
	if id != uuid.Nil {
		t.Error("expected nil UUID")
	}
}

func TestGetUserID_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", http.NoBody)

	c.Set(string(KeyUserID), "not-a-uuid")

	id, ok := GetUserID(c)
	if ok {
		t.Error("expected ok=false for wrong type")
	}
	if id != uuid.Nil {
		t.Error("expected nil UUID")
	}
}

func TestContextKeys(t *testing.T) {
	if KeyUserID == "" {
		t.Error("KeyUserID should not be empty")
	}
	if KeyRequestID == "" {
		t.Error("KeyRequestID should not be empty")
	}
	if KeyUserID == KeyRequestID {
		t.Error("context keys should be different")
	}
}
