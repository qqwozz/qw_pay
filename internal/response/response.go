package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	OK    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Meta  *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{OK: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{OK: true, Data: data})
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{OK: false, Error: msg})
}

func Paginated(c *gin.Context, data interface{}, page, perPage, total int) {
	c.JSON(http.StatusOK, Response{
		OK:   true,
		Data: data,
		Meta: &Meta{Page: page, PerPage: perPage, Total: total},
	})
}
