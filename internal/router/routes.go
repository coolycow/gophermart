package router

import (
	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/handler"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

func setupURLRoutes(
	r *gin.Engine,
	cfg *config.Config,
	repo repository.Repository,
) {
	userService := service.NewUserService(cfg, repo)
	orderService := service.NewOrderService(cfg, repo)

	// Регистрация пользователя (JSON)
	r.POST("/api/user/register", middleware.ContentTypeJSON(), handler.RegisterHandler(userService))

	// Аутентификация пользователя (JSON)
	r.POST("/api/user/login", middleware.ContentTypeJSON(), handler.LoginHandler(userService))

	// Группа маршрутов с обязательной аутентификацией
	authGroup := r.Group("/")
	authGroup.Use(middleware.RequiredAuthMiddleware(userService))

	// Загрузка номера заказа (TEXT/PLAIN)
	authGroup.POST("/api/user/orders", middleware.ContentTextPlainJSON(), handler.PostOrdersHandler(orderService))

	// Получение списка загруженных номеров заказов
	authGroup.GET("/api/user/orders", handler.GetOrdersHandler(orderService))

	// Получение текущего баланса пользователя
	authGroup.GET("/api/user/balance", handler.GetBalanceHandler())

	// Запрос на списание средств
	authGroup.POST("/api/user/balance/withdraw", handler.PostBalanceWithdrawHandler())

	// Получение информации о выводе средств
	authGroup.GET("/api/user/withdrawals", handler.GetWithdrawalsHandler())

	// Получение информации о расчёте начислений баллов лояльности
	authGroup.GET("/api/orders/{number}", handler.GetOrderInfoHandler())
}
