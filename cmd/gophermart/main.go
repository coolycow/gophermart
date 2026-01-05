package main

import (
	"log"

	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/router"
	"go.uber.org/zap"
)

func main() {
	// Инициализируем настройки (приоритет: окружение, флаги, дефолт)
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
		return
	}

	// Выводим настройки в консоль для наглядности
	cfg.PrintConfig()

	// Инициализируем логгер
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Инициализируем репозиторий
	var repo repository.Repository
	repo, err = repository.NewPostgresRepository(cfg.DatabaseURI)

	if err != nil {
		log.Fatalf("Failed to initialize postgres repository: %v", err)
		return
	}

	logger.Log.Info("Initialized postgres repository", zap.String("dsn", cfg.DatabaseURI))

	if cfg.RunMigrations {
		if err = repo.RunMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		logger.Log.Info("Run migrations succeeded")
		return
	}

	// В конце работы приложения необходимо правильно закрыть хранилище.
	defer func() {
		if err = repo.Close(); err != nil {
			logger.Log.Error("Error closing repository", zap.Error(err))
		}
	}()

	// Инициализируем роутер
	r := router.NewRouter(cfg, repo)

	// Получаем адрес сервера из настроек и запускаем сервер
	logger.Log.Info("Running server ", zap.String("address", cfg.RunAddress))
	err = r.Run(cfg.RunAddress)

	// Если сервер не стартовал - фатальная ошибка
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
