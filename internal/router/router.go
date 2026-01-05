package router

import (
	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, repo repository.Repository) *gin.Engine {
	router := gin.Default()

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestGzip())

	setupURLRoutes(router, cfg, repo)

	return router
}
