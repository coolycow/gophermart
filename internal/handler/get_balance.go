package handler

import (
	"net/http"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

// GetBalanceHandler текущий баланс пользователя
func GetBalanceHandler(srv service.BalanceTransactionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Start GetBalanceHandler")

		// Получаем ID пользователя из полученных кук
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			logger.Log.Debug("UserID not found in context")
			_ = c.Error(err)
			return
		}

		balance, err := srv.GetBalanceByUserID(c.Request.Context(), userID)
		if err != nil {
			logger.Log.Debug("Error in GetBalanceHandler")
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, balance)
	}
}
