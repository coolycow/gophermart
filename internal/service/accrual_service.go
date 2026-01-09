package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/coolycow/gophermart/internal/config"
	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
)

// AccrualService Сервис для работы с системой начисления баллов
type AccrualService interface {
	GetAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error)
}

// Реализация сервисного слоя
type accrualService struct {
	repo repository.Repository
	cfg  *config.Config
}

// NewAccrualService инициализация сервиса
func NewAccrualService(cfg *config.Config, repo repository.Repository) AccrualService {
	return &accrualService{
		repo: repo,
		cfg:  cfg,
	}
}

// GetAccrual получить данные по начислению баллов для указанного заказа
func (s *accrualService) GetAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	// Создаем дочерний контекст с таймаутом, если его еще нет в родительском контексте.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.cfg.AccrualSystemAddress+"/api/orders/"+orderNumber, nil)

	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Создаём клиент без жесткого таймаута, так как таймаут управляется контекстом
	client := http.Client{}

	response, err := client.Do(req)
	if err != nil {
		// Проверяем, не истек ли таймаут контекста
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, httpError.CustomError{
				Message:    "request timeout",
				StatusCode: http.StatusRequestTimeout,
			}
		}

		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, httpError.CustomError{
			Message:    http.StatusText(response.StatusCode),
			StatusCode: http.StatusBadRequest,
		}
	}

	var accrual model.Accrual
	err = json.NewDecoder(response.Body).Decode(&accrual)

	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &accrual, nil
}
