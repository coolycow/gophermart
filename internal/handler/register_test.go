package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginHandlerIntegration тестирование логина
func TestRegisterHandlerIntegration(t *testing.T) {
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
			name:     "successful register",
			login:    "new_user",
			password: "new_pass",
			wantCode: http.StatusOK,
		},
		{
			name:     "user already exists",
			login:    "test_user",
			password: "wrong_pass",
			wantCode: http.StatusConflict,
		},
		{
			name:     "empty login",
			login:    "",
			password: "any_pass",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "empty password",
			login:    "any_user",
			password: "",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "login too short",
			login:    "ab",
			password: "valid_password",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "password too short",
			login:    "valid_login",
			password: "ab",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "login too long",
			login:    strings.Repeat("a", 256),
			password: "valid_password",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "password too long",
			login:    "valid_login",
			password: strings.Repeat("a", 256),
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequestGzip())
			router.POST("/api/user/register", RegisterHandler(userService))

			jsonData := fmt.Sprintf(`{"login": "%s", "password": "%s"}`, tt.login, tt.password)
			request := httptest.NewRequest("POST", "/api/user/register", strings.NewReader(jsonData))
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
