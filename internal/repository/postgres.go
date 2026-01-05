package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
)

// PostgresRepository представляет репозиторий для хранения URL
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

	// Указываем путь к директории с миграциями
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx",
		driver,
	)
	if err != nil {
		return err
	}

	// Применяем миграции
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// NewPostgresRepository создает новый экземпляр URLRepository
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

	// Проверяем, существует ли таблица urls
	tableExists, err := repo.checkTableExists("urls")
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
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID int) (model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, login, password, created_at, updated_at, deleted_at from users where id = $1", userID)

	var ID int
	var login, password string
	var createdAt, updatedAt, deletedAt time.Time
	err := row.Scan(&ID, &login, &password, &createdAt, &updatedAt, &deletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, nil
		}

		return model.User{}, err
	}

	return model.User{
		ID:        ID,
		Login:     login,
		Password:  password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}

// GetUserByLogin возвращает пользователя по его ID
func (r *PostgresRepository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, password, created_at, updated_at, deleted_at from users where login = $1", login)

	var ID int
	var password string
	var createdAt, updatedAt, deletedAt time.Time
	err := row.Scan(&ID, &password, &createdAt, &updatedAt, &deletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, nil
		}

		return model.User{}, err
	}

	return model.User{
		ID:        ID,
		Login:     login,
		Password:  password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}, nil
}

// CreateUser создание нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, login string, password string) (model.User, error) {
	row := r.db.QueryRowContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id, created_at, updated_at", login, password)

	var ID int
	var createdAt, updatedAt time.Time
	err := row.Scan(&ID, &createdAt, &updatedAt)

	if err != nil {
		return model.User{}, err
	}

	fmt.Println("CREATE USER WITH ID = ", ID)

	return model.User{ID: ID, CreatedAt: createdAt, UpdatedAt: updatedAt}, nil
}

// DeleteUser удаление пользователя по его ID
func (r *PostgresRepository) DeleteUser(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)

	if err != nil {
		return err
	}

	return nil
}
