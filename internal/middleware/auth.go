package middleware

import (
	"net/http"

	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

type ginKey string

const (
	UserIDKey ginKey = "userID"
)

// GetUserIDFromGinContext возвращает ID пользователя из контекста
func GetUserIDFromGinContext(c *gin.Context) (int, error) {
	value, exists := c.Get(string(UserIDKey))

	if !exists || value == nil {
		return 0, httpError.HTTPError{
			Message:    "user id not found in context",
			StatusCode: http.StatusInternalServerError,
		}
	}

	userID, ok := value.(int)

	if !ok {
		return 0, httpError.HTTPError{
			Message:    "incorrect user id in context",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return userID, nil
}

// RequiredAuthMiddleware проверяет наличие авторизационной куки и записывает ID пользователя в данные контекста
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
			_ = c.Error(httpError.HTTPError{
				Message:    "User with this ID does not exist",
				StatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		if user.DeletedAt != nil {
			_ = c.Error(httpError.HTTPError{
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
