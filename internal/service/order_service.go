package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"runtime"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
	"go.uber.org/zap"

	httpError "github.com/coolycow/gophermart/internal/error"
)

// OrderService Сервис для работы в Handler
type OrderService interface {
	SanitizeOrderNumber(orderNumber string) string
	IsCorrectOrderNumber(orderNumber string) bool

	CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, bool, error)

	GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error)
	GetOrdersForUpdate(ctx context.Context) ([]model.Order, error)

	UpdateOrderTask(ctx context.Context)
	UpdateOrderByAccrual(ctx context.Context, userID int, accrual model.Accrual) error
	UpdateOrderStatusAndAccrual(ctx context.Context, userID int, orderNumber string, status string, accrual float32) error
	ResetStuckProcessingOrders(ctx context.Context) error
}

// Реализация сервисного слоя
type orderService struct {
	repo repository.Repository
	cfg  *config.Config
	acc  AccrualService
}

// NewOrderService инициализация сервиса
func NewOrderService(cfg *config.Config, repo repository.Repository) OrderService {
	return &orderService{
		repo: repo,
		cfg:  cfg,
		acc:  NewAccrualService(cfg, repo),
	}
}

// SanitizeOrderNumber очищает номер заказа от лишних символов
func (s *orderService) SanitizeOrderNumber(orderNumber string) string {
	reg := regexp.MustCompile("[^0-9]+")
	return reg.ReplaceAllString(orderNumber, "")
}

// IsCorrectOrderNumber - проверка правильности номера заказа по алгоритму Луна
func (s *orderService) IsCorrectOrderNumber(orderNumber string) bool {
	return goluhn.Validate(orderNumber) == nil
}

// GetOrdersByUserID возвращает все заказы пользователя по его ID
func (s *orderService) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return orders, nil
}

// GetOrdersForUpdate возвращает все заказы, которые требуют обновления
func (s *orderService) GetOrdersForUpdate(ctx context.Context) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersForUpdate(ctx)
	if err != nil {
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return orders, nil
}

// CreateOrder - добавление нового заказа для пользователя
func (s *orderService) CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, bool, error) {
	clearOrderNumber := s.SanitizeOrderNumber(orderNumber)

	if !s.IsCorrectOrderNumber(clearOrderNumber) {
		return nil, false, httpError.HTTPError{
			Message:    "Invalid order number",
			StatusCode: http.StatusUnprocessableEntity,
		}
	}

	existedOrder, err := s.repo.GetOrderByNumber(ctx, clearOrderNumber)
	if err != nil {
		return nil, false, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if existedOrder != nil {
		if existedOrder.UserID != userID {
			return nil, false, httpError.HTTPError{
				Message:    "order already exists",
				StatusCode: http.StatusConflict,
			}
		}
		return existedOrder, false, nil
	}

	newOrder, err := s.repo.CreateOrder(ctx, userID, clearOrderNumber)
	if err != nil {
		return nil, false, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return newOrder, true, nil
}

// updateOrderWorker обновляет указанный заказ (непосредственно сам Worker)
func updateOrderWorker(ctx context.Context, srv *orderService, id int, orders <-chan model.Order) {
	for order := range orders {
		logger.Log.Debug(fmt.Sprintf("worker %d start update order %s by accrual", id, order.Number))

		accrual, err := srv.acc.GetAccrual(ctx, order.Number)

		if err != nil {
			logger.Log.Error("Get accrual error", zap.Error(err))
			continue
		}

		err = srv.UpdateOrderByAccrual(ctx, order.UserID, *accrual)

		if err != nil {
			logger.Log.Error("Update order by accrual error", zap.Error(err))
			continue
		}

		logger.Log.Debug(fmt.Sprintf("worker %d end update order %s by accrual", id, order.Number))
	}
}

// UpdateOrderTask Обновление статуса и суммы баллов с использованием Worker Pool
func (s *orderService) UpdateOrderTask(ctx context.Context) {
	// Получаем заказы, которые требуется обновить
	orders, err := s.repo.GetOrdersForUpdate(ctx)

	if err != nil {
		logger.Log.Error("Update order task error", zap.Error(err))
		return
	}

	if len(orders) == 0 {
		logger.Log.Debug("Order list for update accrual is empty")
		return
	}

	// Количество задач - это количество заказов для обновления
	numJobs := len(orders)

	// Количество воркеров задаём по количеству ядер процессора
	numWorkers := runtime.NumCPU()

	logger.Log.Debug(fmt.Sprintf("Start update accrual for %d orders with %d workers", numJobs, numWorkers))

	// Создаем буферизованный канал для принятия задач в воркер
	jobs := make(chan model.Order, numJobs)

	// Создаем и запускаем numWorkers воркеров - это будет наш пул
	for w := 1; w <= numWorkers; w++ {
		go updateOrderWorker(ctx, s, w, jobs)
	}

	// В канал задач отправляем заказы
	for j := 1; j <= numJobs; j++ {
		jobs <- orders[j-1]
	}

	// Закрываем канал на стороне отправителя
	close(jobs)
}

// UpdateOrderByAccrual обновление статуса заказа и суммы начислений на основе объекта модели Accrual
func (s *orderService) UpdateOrderByAccrual(ctx context.Context, userID int, accrual model.Accrual) error {
	return s.UpdateOrderStatusAndAccrual(ctx, userID, accrual.Order, accrual.Status, accrual.Accrual)
}

// UpdateOrderStatusAndAccrual обновление статуса заказа и суммы начислений на основе отдельных полей
func (s *orderService) UpdateOrderStatusAndAccrual(ctx context.Context, userID int, orderNumber string, status string, accrual float32) error {
	return s.repo.UpdateOrderStatusAndAccrual(ctx, userID, orderNumber, status, accrual)
}

// ResetStuckProcessingOrders сбрасывает статус зависших заказов с PROCESSING на NEW
func (s *orderService) ResetStuckProcessingOrders(ctx context.Context) error {
	return s.repo.ResetStuckProcessingOrders(ctx)
}
