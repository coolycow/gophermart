package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabase структура для хранения информации о тестовой БД
type TestDatabase struct {
	DSN      string
	Name     string
	ConnPool *sql.DB
}

// setupTestDB создаёт и настраивает тестовую базу данных
func setupTestDB(t *testing.T) *TestDatabase {
	t.Helper()

	// Параметры подключения к PostgreSQL (адаптируйте под вашу систему)
	host := "localhost"
	port := "5433"
	user := "postgres"
	password := "RriExM4l6CO6NBfGSp7E" // или ваш пароль

	// Создаём уникальное имя для тестовой БД
	dbName := fmt.Sprintf("gophermart_test_%d_%d", time.Now().Unix(), os.Getpid())

	// Сначала подключаемся к PostgreSQL без указания БД (к системной БД postgres)
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		user, password, host, port)

	db, err := sql.Open("pgx", adminDSN)
	require.NoError(t, err)
	defer db.Close()

	// Создаём тестовую базу данных
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	// Формируем DSN для подключения к созданной тестовой БД
	testDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName)

	// Подключаемся к тестовой БД
	testDB, err := sql.Open("pgx", testDSN)
	require.NoError(t, err)

	// Проверяем соединение
	err = testDB.Ping()
	require.NoError(t, err)

	return &TestDatabase{
		DSN:      testDSN,
		Name:     dbName,
		ConnPool: testDB,
	}
}

// cleanupTestDB удаляет тестовую базу данных
func cleanupTestDB(t *testing.T, testDB *TestDatabase) {
	t.Helper()

	if testDB.ConnPool != nil {
		testDB.ConnPool.Close()
	}

	// Подключаемся к системной БД для удаления тестовой БД
	host := "localhost"
	port := "5433"
	user := "postgres"
	password := "RriExM4l6CO6NBfGSp7E"

	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
		user, password, host, port)

	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Logf("Failed to connect to admin DB for cleanup: %v", err)
		return
	}
	defer db.Close()

	// Удаляем тестовую БД
	// Сначала отключаем все активные соединения
	_, err = db.Exec(fmt.Sprintf(`
        SELECT pg_terminate_backend(pg_stat_activity.pid) 
        FROM pg_stat_activity 
        WHERE pg_stat_activity.datname = '%s' 
          AND pid <> pg_backend_pid()`, testDB.Name))

	if err != nil {
		t.Logf("Failed to terminate connections: %v", err)
	}

	// Теперь удаляем БД
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", testDB.Name))
	if err != nil {
		t.Logf("Failed to drop test database: %v", err)
	}
}

// createTestUser создаёт тестового пользователя в БД
func createTestUser(t *testing.T, srv service.UserService, login string, password string) *model.User {
	t.Helper()

	// Создаём пользователя
	user, err := srv.CreateUser(context.Background(), model.UserRegister{Login: login, Password: password})
	require.NoError(t, err)

	return user
}

// TestLoginHandlerIntegration тестирование логина
func TestLoginHandlerIntegration(t *testing.T) {
	// Настраиваем тестовую БД
	testDB := setupTestDB(t)
	defer cleanupTestDB(t, testDB)

	// Создаём репозиторий с тестовой БД
	repo, err := repository.NewPostgresRepository(testDB.DSN)
	require.NoError(t, err)
	defer repo.Close()

	// Настраиваем сервис для работы с пользователем
	cfg, _ := config.InitConfig()
	userService := service.NewUserService(cfg, repo)

	// Создаём тестового пользователя
	_ = createTestUser(t, userService, "test_user", "test_pass")

	tests := []struct {
		name     string
		login    string
		password string
		wantCode int
	}{
		{
			name:     "successful login",
			login:    "test_user",
			password: "test_pass",
			wantCode: http.StatusOK,
		},
		{
			name:     "wrong password",
			login:    "test_user",
			password: "wrong_pass",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "user not found",
			login:    "nonexistent",
			password: "any_pass",
			wantCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequestGzip())
			router.POST("/api/user/login", LoginHandler(userService))

			jsonData := fmt.Sprintf(`{"login": "%s", "password": "%s"}`, tt.login, tt.password)
			request := httptest.NewRequest("POST", "/api/user/login", strings.NewReader(jsonData))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			assert.Equal(t, tt.wantCode, w.Code)

			if tt.wantCode == http.StatusOK {
				// Проверяем наличие куки
				cookies := w.Result().Cookies()
				var authCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "auth" {
						authCookie = c
						break
					}
				}
				assert.NotNil(t, authCookie)
				assert.True(t, authCookie.HttpOnly)
				assert.Equal(t, "/", authCookie.Path)
			}
		})
	}
}
