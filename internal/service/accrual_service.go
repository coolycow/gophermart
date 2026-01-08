package service

import (
	"context"
	"encoding/json"
	"net/http"

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

func (s *accrualService) GetAccrual(ctx context.Context, orderNumber string) (*model.Accrual, error) {
	response, err := http.Get(s.cfg.AccrualSystemAddress + "/api/orders/" + orderNumber)

	if err != nil {
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
