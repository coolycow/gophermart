package middleware

import (
	"net/http"
	"strings"

	"github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/gin-gonic/gin"
)

// ContentTypeJSON проверяем, что тип контента - application/json
func ContentTypeJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			logger.Log.Debug("content type is not application/json")

			_ = c.Error(error.CustomError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})

			c.Abort()

			return
		}

		c.Next()
	}
}
