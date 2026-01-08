package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresRepository представляет репозиторий для хранения
type PostgresRepository struct {
	db *sql.DB
}

// checkTableExists проверяет, существует ли таблица
func (r *PostgresRepository) checkTableExists(tableName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`, tableName).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) RunMigrations() error {
	logger.Log.Info("Running migrations")

	// Создаем экземпляр драйвера для PostgreSQL
	driver, err := pgx.WithInstance(r.db, &pgx.Config{})
	if err != nil {
		return err
	}

	// Получаем абсолютный путь к директории с миграциями
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ищем корень проекта (где находится go.mod)
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}

		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			return fmt.Errorf("project root not found")
		}
		projectRoot = parent
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")
	migrationsURL := "file://" + filepath.ToSlash(migrationsPath)

	// Указываем путь к директории с миграциями
	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "pgx", driver)
	if err != nil {
		return err
	}

	// Применяем миграции
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// NewPostgresRepository создает новый экземпляр Repository
func NewPostgresRepository(DSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
		return nil, err
	}

	repo := &PostgresRepository{db: db}

	// Проверяем, существует ли таблица users
	tableExists, err := repo.checkTableExists("users")
	if err != nil {
		return nil, err
	}

	if !tableExists {
		err = repo.RunMigrations()

		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

// Close закрывает хранилище
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Ping проверяет доступность хранилища
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// GetUserByID возвращает пользователя по его ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, login, password, created_at, updated_at, deleted_at from users where id = $1", userID)

	var ID int
	var login, password string
	var createdAt, updatedAt, deletedAt *time.Time
	err := row.Scan(&ID, &login, &password, &createdAt, &updatedAt, &deletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &model.User{
		ID:        ID,
		Login:     login,
		Password:  password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}

// GetUserByLogin возвращает пользователя по его ID
func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, password, created_at, updated_at, deleted_at from users where login = $1", login)

	var ID int
	var password string
	var createdAt, updatedAt, deletedAt *time.Time
	err := row.Scan(&ID, &password, &createdAt, &updatedAt, &deletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &model.User{
		ID:        ID,
		Login:     login,
		Password:  password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}

// CreateUser создание нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, login string, password string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id, created_at, updated_at", login, password)

	var ID int
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &model.User{ID: ID, Login: login, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

// DeleteUser удаление пользователя по его ID
func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

	if err != nil {
		return err
	}

	return nil
}

// GetOrderByNumber возвращает заказ по его номеру
func (r *PostgresRepository) GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	row := r.db.QueryRowContext(ctx, "select id, user_id, status, created_at, updated_at from orders where number = $1", orderNumber)

	var ID, userId int
	var status string
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &userId, &status, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &model.Order{
		ID:        ID,
		UserID:    userId,
		Number:    orderNumber,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetOrderByUserIDNumber возвращает заказ по пользователю и номеру заказа
func (r *PostgresRepository) GetOrderByUserIDNumber(ctx context.Context, userID int, orderNumber string) (*model.Order, error) {
	row := r.db.QueryRowContext(ctx, "select id, created_at, status, updated_at from orders where user_id = $1 AND number = $2", userID, orderNumber)

	var ID int
	var status string
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &status, &createdAt, &updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &model.Order{
		ID:        ID,
		UserID:    userID,
		Number:    orderNumber,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetOrdersByUserID возвращает все заказы пользователя по его ID
func (r *PostgresRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	var result []model.Order

	rows, err := r.db.QueryContext(ctx, "select id, number, status, created_at, updated_at from orders where user_id = $1 ORDER BY created_at DESC", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var ID int
		var number, status string
		var createdAt, updatedAt *time.Time
		err := rows.Scan(&ID, &number, &status, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, model.Order{
			ID:        ID,
			Number:    number,
			UserID:    userID,
			Status:    status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateOrder создание нового заказа с привязкой к пользователю
func (r *PostgresRepository) CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, error) {
	row := r.db.QueryRowContext(ctx, "INSERT INTO orders (user_id, number) VALUES ($1, $2) RETURNING id, status, created_at, updated_at", userID, orderNumber)

	var ID int
	var status string
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &status, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:        ID,
		UserID:    userID,
		Number:    orderNumber,
		Status:    status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
