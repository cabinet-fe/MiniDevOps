package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API envelope { code, message, data? }.
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "created", Data: data})
}

func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{Code: httpCode, Message: message})
	c.Abort()
}
