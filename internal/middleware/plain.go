package middleware

import (
	"net/http"
	"strings"

	"github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/gin-gonic/gin"
)

// ContentTextPlainJSON проверяем, что тип контента - text/plain
func ContentTextPlainJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			logger.Log.Debug("content type is not text/plain")
			_ = c.Error(error.HttpError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})

			c.Abort()

			return
		}

		c.Next()
	}
}
