package handler

import (
	"encoding/json"
	"net/http"

	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostBalanceWithdrawHandler - запрос на списание средств
func PostBalanceWithdrawHandler(srv service.BalanceTransactionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Start PostBalanceWithdrawHandler")

		// Получаем ID пользователя из полученных кук
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			logger.Log.Debug("UserID not found in context")
			_ = c.Error(err)
			return
		}

		// Получаем данные пользователя из запроса
		var req model.WithdrawalRequest
		dec := json.NewDecoder(c.Request.Body)

		if err = dec.Decode(&req); err != nil {
			logger.Log.Debug("Cannot decode request JSON body", zap.Error(err))
			_ = c.Error(httpError.HttpError{
				Message:    "Cannot decode request JSON body",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Пробуем списать средства
		_, err = srv.CreateWithdraw(c.Request.Context(), userID, req.Order, req.Sum)
		if err != nil {
			logger.Log.Debug("Cannot create withdraw", zap.Error(err))
			_ = c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	}
}
