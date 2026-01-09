package handler

import (
	"net/http"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

// GetWithdrawalsHandler - получение информации о выводе средств
func GetWithdrawalsHandler(srv service.BalanceTransactionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Start GetWithdrawalsHandler")

		// Получаем ID пользователя из полученных кук
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			logger.Log.Debug("UserID not found in context")
			_ = c.Error(err)
			return
		}

		balanceTransactions, err := srv.GetWithdrawalsByUserID(c.Request.Context(), userID)
		if err != nil {
			logger.Log.Debug("Error in GetWithdrawalsHandler")
			_ = c.Error(err)
			return
		}

		if len(balanceTransactions) == 0 {
			logger.Log.Debug("No balance withdrawal found")
			c.Status(http.StatusNoContent)
			return
		}

		var result []model.WithdrawalResponse

		for _, balanceTransaction := range balanceTransactions {
			result = append(result, balanceTransaction.ToWithdrawalResponse())
		}

		c.JSON(http.StatusOK, result)
	}
}
