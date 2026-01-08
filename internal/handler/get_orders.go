package handler

import (
	"net/http"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

// GetOrdersHandler возвращает все заказы пользователя
func GetOrdersHandler(orderService service.OrderService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Start GetOrdersHandler")

		// Получаем ID пользователя из полученных кук
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			logger.Log.Debug("UserID not found in context")
			_ = c.Error(err)
			return
		}

		// Получаем заказы пользователя
		orders, err := orderService.GetOrdersByUserID(c.Request.Context(), userID)

		if err != nil {
			logger.Log.Debug("Error get orders list from DB")
			_ = c.Error(err)
			return
		}

		if orders == nil || len(orders) == 0 {
			logger.Log.Debug("Orders list is empty")
			c.Status(http.StatusNoContent)
			return
		}

		// Получаем список заказов в нужном формате
		var ordersList []model.OrderResponse
		for _, order := range orders {
			ordersList = append(ordersList, order.ToResponse())
		}

		c.JSON(http.StatusOK, ordersList)
	}
}
