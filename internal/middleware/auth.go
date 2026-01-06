package middleware

import (
	"errors"
	"net/http"

	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

type ginKey string

const (
	UserIDKey ginKey = "userID"
)

// GetUserIDFromGinContext возвращает ID пользователя
func GetUserIDFromGinContext(c *gin.Context) (string, error) {
	value, exists := c.Get(string(UserIDKey))

	if !exists || value == nil {
		return "", errors.New("user ID not found in context")
	}

	userID, ok := value.(string)

	if !ok {
		return "", errors.New("incorrect user ID in context")
	}

	return userID, nil
}

// RequiredAuthMiddleware проверяет наличие авторизационной куки
func RequiredAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userID, err := userService.GetUserIDFromCookie(cookie)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid cookie"})
			c.Abort()
			return
		}

		user, err := userService.GetUserByID(c.Request.Context(), userID)

		if err != nil {
			_ = c.Error(httpError.CustomError{
				Message:    "User with this ID does not exist",
				StatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		if !user.DeletedAt.IsZero() {
			_ = c.Error(httpError.CustomError{
				Message:    "User with this ID has already been deleted",
				StatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), userID)

		c.Next()
	}
}
