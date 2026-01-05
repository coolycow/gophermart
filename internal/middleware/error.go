package middleware

import (
	"errors"

	"github.com/coolycow/gophermart/internal/error"
	"github.com/gin-gonic/gin"
)

// ErrorHandler captures error and returns a consistent JSON error response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				var customErr error.CustomError
				if errors.As(ginErr.Err, &customErr) {
					c.JSON(customErr.StatusCode, gin.H{"error": customErr.Message})
					return
				}
			}
		}
	}
}
