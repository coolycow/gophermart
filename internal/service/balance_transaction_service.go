package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/coolycow/gophermart/internal/config"
	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
)

// BalanceTransactionService Сервис для работы с системой начисления баллов
type BalanceTransactionService interface {
	GetBalanceByUserID(ctx context.Context, userID int) (*model.BalanceResponse, error)
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]model.BalanceTransaction, error)

	CreateAccrual(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error)
	CreateWithdraw(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error)
}

// Реализация сервисного слоя
type balanceTransactionService struct {
	repo         repository.Repository
	cfg          *config.Config
	orderService OrderService
}

// NewBalanceTransactionService инициализация сервиса
func NewBalanceTransactionService(cfg *config.Config, repo repository.Repository) BalanceTransactionService {
	return &balanceTransactionService{
		repo:         repo,
		cfg:          cfg,
		orderService: NewOrderService(cfg, repo),
	}
}

// GetBalanceByUserID возвращает текущий баланс пользователя
func (s *balanceTransactionService) GetBalanceByUserID(ctx context.Context, userID int) (*model.BalanceResponse, error) {
	balance, err := s.repo.GetBalanceByUserID(ctx, userID)

	if err != nil {
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return balance, nil
}

// GetWithdrawalsByUserID возвращает
func (s *balanceTransactionService) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]model.BalanceTransaction, error) {
	balanceTransactions, err := s.repo.GetWithdrawalsByUserID(ctx, userID)

	if err != nil {
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return balanceTransactions, nil
}

// CreateAccrual создаёт запись об изменении баланса
func (s *balanceTransactionService) CreateAccrual(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error) {
	clearOrderNumber := s.orderService.SanitizeOrderNumber(orderNumber)

	// Проверяем номер заказа на корректность
	if !s.orderService.IsCorrectOrderNumber(clearOrderNumber) {
		return nil, httpError.HTTPError{
			Message:    "Invalid order number",
			StatusCode: http.StatusInternalServerError,
		}
	}

	balance, err := s.repo.CreateAccrual(ctx, userID, clearOrderNumber, amount)

	if err != nil {
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return balance, nil
}

// CreateWithdraw создаёт запись о списании средств
func (s *balanceTransactionService) CreateWithdraw(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error) {
	clearOrderNumber := s.orderService.SanitizeOrderNumber(orderNumber)

	// Проверяем номер заказа на корректность
	if !s.orderService.IsCorrectOrderNumber(clearOrderNumber) {
		return nil, httpError.HTTPError{
			Message:    "Invalid order number",
			StatusCode: http.StatusUnprocessableEntity,
		}
	}

	balance, err := s.repo.CreateWithdraw(ctx, userID, clearOrderNumber, amount)
	if err != nil {
		var insufficientBalanceErr *httpError.InsufficientBalanceError

		if errors.As(err, &insufficientBalanceErr) {
			return nil, httpError.HTTPError{
				Message:    err.Error(),
				StatusCode: insufficientBalanceErr.StatusCode(),
			}
		}
		return nil, httpError.HTTPError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return balance, nil
}
