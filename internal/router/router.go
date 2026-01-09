package router

import (
	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

// NewRouter - создание нового роутера
func NewRouter(cfg *config.Config, repo repository.Repository) *gin.Engine {
	router := gin.Default()

	// Настраиваем поддержку сжатых данных
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Подключаем нужные middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestGzip())

	// Настраиваем все необходимые маршруты
	setupURLRoutes(router, cfg, repo)

	return router
}
