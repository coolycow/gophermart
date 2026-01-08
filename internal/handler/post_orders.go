package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

func getTextPlainContent(c *gin.Context) (string, error) {
	// Читаем тело запроса
	body, err := io.ReadAll(c.Request.Body)

	if err != nil {
		return "", httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Проверяем, что не пришла пустота
	if len(body) == 0 {
		return "", httpError.CustomError{
			Message:    "Empty body",
			StatusCode: http.StatusBadRequest,
		}
	}

	// Извлекаем строку из тела запроса и проверяем, что она не пуста
	trimBody := strings.TrimSpace(string(body))

	if trimBody == "" {
		return "", httpError.CustomError{
			Message:    "Empty Content",
			StatusCode: http.StatusBadRequest,
		}
	}

	return trimBody, nil
}

// PostOrdersHandler - загрузка номера заказа
func PostOrdersHandler(orderService service.OrderService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("Start post orders handler")

		// Читаем тело запроса
		content, err := getTextPlainContent(c)

		if err != nil {
			_ = c.Error(err)
			return
		}

		// Получаем номер заказа
		orderNumber := orderService.ClearOrderNumber(content)
		logger.Log.Debug("Clear order number: " + orderNumber)

		// Проверяем корректность номера заказа по алгоритму Луна
		if !orderService.IsCorrectOrderNumber(orderNumber) {
			logger.Log.Debug("Order number " + orderNumber + " is not correct")
			_ = c.Error(httpError.CustomError{
				Message:    "Invalid order number",
				StatusCode: http.StatusBadRequest,
			})
			return
		}
		logger.Log.Debug("Order number " + orderNumber + " is correct")

		// Получаем ID пользователя из полученных кук
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			logger.Log.Debug("UserID not found in context")
			_ = c.Error(err)
			return
		}
		logger.Log.Debug("User ID: " + strconv.Itoa(userID))

		// Добавляем заказ для пользователя
		order, isNew, err := orderService.CreateOrder(c.Request.Context(), userID, orderNumber)
		if err != nil {
			logger.Log.Debug("Error creating new order")
			_ = c.Error(httpError.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		if isNew {
			logger.Log.Debug("New order " + orderNumber + " created successfully with ID: " + strconv.Itoa(order.ID))
			c.Status(http.StatusCreated)
		} else {
			logger.Log.Debug("Order " + orderNumber + " was created earlier with ID: " + strconv.Itoa(order.ID))
			c.Status(http.StatusOK)
		}
	}
}
