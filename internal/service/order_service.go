package service

import (
	"context"
	"net/http"
	"regexp"

	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"

	httpError "github.com/coolycow/gophermart/internal/error"
)

// OrderService Сервис для работы в Handler
type OrderService interface {
	ClearOrderNumber(orderNumber string) string
	IsCorrectOrderNumber(orderNumber string) bool
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, bool, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error)

	UpdateOrderByAccrual(ctx context.Context, accrual model.Accrual) error
	UpdateOrderStatusAndAccrual(ctx context.Context, orderNumber string, status string, accrual float32) error
}

// Реализация сервисного слоя
type orderService struct {
	repo repository.Repository
	cfg  *config.Config
}

// NewOrderService инициализация сервиса
func NewOrderService(cfg *config.Config, repo repository.Repository) OrderService {
	return &orderService{
		repo: repo,
		cfg:  cfg,
	}
}

// ClearOrderNumber очищает номер заказа от лишних символов
func (s *orderService) ClearOrderNumber(orderNumber string) string {
	reg := regexp.MustCompile("[^0-9]+")
	return reg.ReplaceAllString(orderNumber, "")
}

// IsCorrectOrderNumber - проверка правильности номера заказа по алгоритму Луна
func (s *orderService) IsCorrectOrderNumber(orderNumber string) bool {
	err := goluhn.Validate(orderNumber)
	if err != nil {
		return false
	}

	return true
}

// GetOrdersByUserID возвращает все заказы пользователя по его ID
func (s *orderService) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	orders, err := s.repo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return orders, nil
}

// CreateOrder - добавление нового заказа для пользователя
func (s *orderService) CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, bool, error) {
	clearOrderNumber := s.ClearOrderNumber(orderNumber)

	existedOrder, err := s.repo.GetOrderByNumber(ctx, clearOrderNumber)
	if err != nil {
		return nil, false, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if existedOrder != nil {
		if existedOrder.UserID != userID {
			return nil, false, httpError.CustomError{
				Message:    "order already exists",
				StatusCode: http.StatusConflict,
			}
		}
		return existedOrder, false, nil
	}

	newOrder, err := s.repo.CreateOrder(ctx, userID, clearOrderNumber)
	if err != nil {
		return nil, false, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return newOrder, true, nil
}

// UpdateOrderByAccrual обновление статуса заказа и суммы начислений
func (s *orderService) UpdateOrderByAccrual(ctx context.Context, accrual model.Accrual) error {
	return s.UpdateOrderStatusAndAccrual(ctx, accrual.Order, accrual.Status, accrual.Accrual)
}

// UpdateOrderStatusAndAccrual обновление статуса заказа и суммы начислений
func (s *orderService) UpdateOrderStatusAndAccrual(ctx context.Context, orderNumber string, status string, accrual float32) error {
	return s.repo.UpdateOrderStatusAndAccrual(ctx, orderNumber, status, accrual)
}
