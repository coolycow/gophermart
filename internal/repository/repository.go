package repository

import (
	"context"

	"github.com/coolycow/gophermart/internal/model"
)

// Repository определяет интерфейс для работы с хранилищем
type Repository interface {
	Close() error
	Ping(ctx context.Context) error
	RunMigrations() error

	GetUserByID(ctx context.Context, userID int) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)

	CreateUser(ctx context.Context, login string, password string) (*model.User, error)
	DeleteUser(ctx context.Context, userID int) error

	GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error)
	GetOrderByUserIDNumber(ctx context.Context, userID int, orderNumber string) (*model.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error)
	GetOrdersForUpdate(ctx context.Context) ([]model.Order, error)

	CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, error)
	UpdateOrderStatusAndAccrual(ctx context.Context, orderNumber string, status string, accrual float32) error
}
