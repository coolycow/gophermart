package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	balanceError "github.com/coolycow/gophermart/internal/error"
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
	row := r.db.QueryRowContext(ctx, `
		SELECT id, login, password, created_at, updated_at, deleted_at 
		FROM users WHERE id = $1`, userID)

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
	row := r.db.QueryRowContext(ctx, `
		SELECT id, password, created_at, updated_at, deleted_at 
		FROM users WHERE login = $1`, login)

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
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO users (login, password) 
		VALUES ($1, $2) 
		RETURNING id, created_at, updated_at`, login, password)

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
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, status, accrual, created_at, updated_at 
		FROM orders WHERE number = $1`, orderNumber)

	var ID, userID int
	var status string
	var accrual float32
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &userID, &status, &accrual, &createdAt, &updatedAt)

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
		Accrual:   accrual,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetOrderByUserIDNumber возвращает заказ по пользователю и номеру заказа
func (r *PostgresRepository) GetOrderByUserIDNumber(ctx context.Context, userID int, orderNumber string) (*model.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, created_at, status, accrual, updated_at 
		FROM orders WHERE user_id = $1 AND number = $2`, userID, orderNumber)

	var ID int
	var status string
	var accrual float32
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &status, &accrual, &createdAt, &updatedAt)

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
		Accrual:   accrual,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// GetOrdersByUserID возвращает все заказы пользователя по его ID
func (r *PostgresRepository) GetOrdersByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	var result []model.Order

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, number, status, accrual, created_at, updated_at 
		FROM orders WHERE user_id = $1 
        ORDER BY created_at DESC`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var ID int
		var accrual float32
		var number, status string
		var createdAt, updatedAt *time.Time
		err = rows.Scan(&ID, &number, &status, &accrual, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, model.Order{
			ID:        ID,
			Number:    number,
			UserID:    userID,
			Status:    status,
			Accrual:   accrual,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetOrdersForUpdate получить список заказов, которые требуют обновления.
// Дополнительно обновляет статус собранных заказов на PROCESSING.
func (r *PostgresRepository) GetOrdersForUpdate(ctx context.Context) ([]model.Order, error) {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var result []model.Order

	// Получаем заказы с блокировкой и сразу обновляем их статус
	rows, err := tx.QueryContext(ctx, `
        SELECT id, user_id, number, status, accrual, created_at, updated_at 
        FROM orders 
        WHERE status IN ('NEW', 'PROCESSING') 
        AND (updated_at < NOW() - INTERVAL '30 seconds' OR status = 'NEW')
        ORDER BY created_at ASC
        FOR UPDATE SKIP LOCKED
        LIMIT 100
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Собираем ID заказов для обновления статуса
	var orderIDs []int

	// Проходим по строкам и собираем result и orderIDs
	for rows.Next() {
		var ID int
		var userID int
		var accrual float32
		var createdAt, updatedAt *time.Time
		var status, number string

		err = rows.Scan(&ID, &userID, &number, &status, &accrual, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, model.Order{
			ID:        ID,
			UserID:    userID,
			Number:    number,
			Status:    status,
			Accrual:   accrual,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})

		orderIDs = append(orderIDs, ID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Если есть заказы, обновляем их статус на PROCESSING
	if len(orderIDs) > 0 {
		placeholders := make([]string, len(orderIDs))
		args := make([]interface{}, len(orderIDs))

		for i, id := range orderIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}

		query := fmt.Sprintf(
			"UPDATE orders SET status = 'PROCESSING', updated_at = NOW() WHERE id IN (%s)",
			strings.Join(placeholders, ","),
		)

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateOrder создание нового заказа с привязкой к пользователю
func (r *PostgresRepository) CreateOrder(ctx context.Context, userID int, orderNumber string) (*model.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, number) 
		VALUES ($1, $2) 
		RETURNING id, status, accrual, created_at, updated_at`, userID, orderNumber)

	var ID int
	var status string
	var accrual float32
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &status, &accrual, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:        ID,
		UserID:    userID,
		Number:    orderNumber,
		Status:    status,
		Accrual:   accrual,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// UpdateOrderStatusAndAccrual обновление статуса заказа и суммы начислений
func (r *PostgresRepository) UpdateOrderStatusAndAccrual(ctx context.Context, userID int, orderNumber string, status string, accrual float32) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET status = $1, accrual = $2, updated_at = NOW() 
		WHERE number = $3`, status, accrual, orderNumber)

	if status == "PROCESSED" && accrual > 0 {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO balance_transactions (user_id, order_number, amount) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`, userID, orderNumber, accrual)

		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// Если произошла ошибка - отменяем транзакцию
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Завершаем транзакцию
	return tx.Commit()
}

// ResetStuckProcessingOrders сбрасывает статус зависших заказов с PROCESSING на NEW
func (r *PostgresRepository) ResetStuckProcessingOrders(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE orders 
        SET status = 'NEW', updated_at = NOW() 
        WHERE status = 'PROCESSING' 
        AND updated_at < NOW() - INTERVAL '5 minutes'
    `)
	return err
}

// GetBalanceByUserID возвращает информацию о текущем балансе пользователя
func (r *PostgresRepository) GetBalanceByUserID(ctx context.Context, userID int) (*model.BalanceResponse, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT  
    	COALESCE(SUM(amount), 0) AS current,
    	COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) AS withdrawn
		FROM balance_transactions WHERE user_id = $1`, userID)

	var current, withdrawn float32
	err := row.Scan(&current, &withdrawn)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &model.BalanceResponse{
		Current:   current,
		Withdrawn: withdrawn,
	}, nil
}

// GetWithdrawalsByUserID возвращает все операции пользователя
func (r *PostgresRepository) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]model.BalanceTransaction, error) {
	var result []model.BalanceTransaction

	rows, err := r.db.QueryContext(ctx, `
			SELECT id, order_number, amount, created_at, updated_at 
			FROM balance_transactions WHERE user_id = $1 AND amount < 0
            ORDER BY created_at DESC`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var ID int
		var amount float32
		var orderNumber string
		var createdAt, updatedAt *time.Time

		err = rows.Scan(&ID, &orderNumber, &amount, &createdAt, &updatedAt)

		if err != nil {
			return nil, err
		}

		result = append(result, model.BalanceTransaction{
			ID:          ID,
			UserID:      userID,
			OrderNumber: orderNumber,
			Amount:      amount,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateAccrual создаёт пополнение в таблице транзакций баллов
func (r *PostgresRepository) CreateAccrual(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO balance_transactions (user_id, order_number, amount) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`, userID, orderNumber, amount)

	var ID int
	var createdAt, updatedAt *time.Time
	err := row.Scan(&ID, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return &model.BalanceTransaction{
		ID:          ID,
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// CreateWithdraw создаёт списание в таблице транзакций баллов
func (r *PostgresRepository) CreateWithdraw(ctx context.Context, userID int, orderNumber string, amount float32) (*model.BalanceTransaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}

	// Сначала блокируем все строки пользователя
	_, err = tx.ExecContext(ctx, `
		SELECT id FROM balance_transactions 
		WHERE user_id = $1 FOR UPDATE`, userID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Затем вычисляем баланс. FOR UPDATE нельзя использовать с агрегатной функцией SUM
	row := tx.QueryRowContext(ctx, `
		SELECT  
		COALESCE(SUM(amount), 0) AS current
		FROM balance_transactions WHERE user_id = $1`, userID)

	var current float32
	err = row.Scan(&current)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if current < amount {
		_ = tx.Rollback()
		return nil, &balanceError.InsufficientBalanceError{Message: "insufficient balance"}
	}

	row = tx.QueryRowContext(ctx, `
		INSERT INTO balance_transactions (user_id, order_number, amount) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`, userID, orderNumber, amount)

	var ID int
	var createdAt, updatedAt *time.Time
	err = row.Scan(&ID, &createdAt, &updatedAt)

	// Если произошла ошибка - отменяем транзакцию
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Завершаем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Завершаем транзакцию
	return &model.BalanceTransaction{
		ID:          ID,
		UserID:      userID,
		OrderNumber: orderNumber,
		Amount:      amount,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
