package main

import (
	"context"
	"log"
	"time"

	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/router"
	"github.com/coolycow/gophermart/internal/service"
	"go.uber.org/zap"
)

// updateOrderTask Обновление заказов
func updateOrderTask(ctx context.Context, srv service.OrderService) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			return

		// выполняем нужный нам код
		default:
			srv.UpdateOrderTask(ctx)
		}

		// делаем паузу перед следующей итерацией
		time.Sleep(5 * time.Second)
	}
}

// resetOrderTask сброс зависших заказов
func resetOrderTask(ctx context.Context, srv service.OrderService) {
	// запускаем бесконечный цикл
	for {
		select {
		// проверяем не завершён ли ещё контекст и выходим, если завершён
		case <-ctx.Done():
			return

		// выполняем нужный нам код
		default:
			_ = srv.ResetStuckProcessingOrders(ctx)
		}

		// делаем паузу перед следующей итерацией
		time.Sleep(5 * time.Minute)
	}
}

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

	// Запускаем миграции если это требуется
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

	// Запускаем очередь заданий на обновление заказов
	orderService := service.NewOrderService(cfg, repo)
	ctxUpdateOrderTask, cancelUpdateOrderTask := context.WithCancel(context.Background())
	go updateOrderTask(ctxUpdateOrderTask, orderService)
	defer cancelUpdateOrderTask()
	logger.Log.Info("Start Update Order Task")

	// Запускаем задание на сброс зависших заказов
	ctxResetOrderTask, cancelResetOrderTask := context.WithCancel(context.Background())
	go resetOrderTask(ctxResetOrderTask, orderService)
	defer cancelResetOrderTask()
	logger.Log.Info("Start Reset Order Task")

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
