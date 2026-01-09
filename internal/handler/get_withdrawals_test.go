package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coolycow/gophermart/internal/config"
	"github.com/coolycow/gophermart/internal/middleware"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostOrdersHandlerIntegration тестирование добавления заказа
func TestGetWithdrawalsHandlerIntegration(t *testing.T) {
	// Настраиваем тестовую БД
	testDB := setupTestDB(t)
	defer cleanupTestDB(t, testDB)

	// Создаём репозиторий с тестовой БД
	repo, err := repository.NewPostgresRepository(testDB.DSN)
	require.NoError(t, err)
	defer repo.Close()

	// Настраиваем сервис для работы с пользователем
	cfg, _ := config.InitConfig()
	cfg.LogLevel = "debug"
	userService := service.NewUserService(cfg, repo)
	balanceTransactionService := service.NewBalanceTransactionService(cfg, repo)

	// Создаём тестовых пользователя
	user1 := createTestUser(t, userService, "test_user_1", "test_pass_1")
	user2 := createTestUser(t, userService, "test_user_2", "test_pass_2")
	user3 := createTestUser(t, userService, "test_user_3", "test_pass_3")

	// Создаём тестовый заказ
	_ = createTestAccrual(t, balanceTransactionService, user1.ID, "5580 4733 7202 4733", 1000)
	_ = createTestWithdraw(t, balanceTransactionService, user1.ID, "5580 4733 7202 4733", 700)

	_ = createTestAccrual(t, balanceTransactionService, user2.ID, "5580 4733 7202 4733", 999.95)

	tests := []struct {
		name          string
		user          *model.User
		withoutCookie bool
		wantCode      int
	}{
		{
			name:     "current and withdrawn",
			user:     user1,
			wantCode: http.StatusOK,
		},
		{
			name:     "only current",
			user:     user2,
			wantCode: http.StatusNoContent,
		},
		{
			name:     "empty balance",
			user:     user3,
			wantCode: http.StatusNoContent,
		},
		{
			name:          "without cookie",
			user:          user1,
			withoutCookie: true,
			wantCode:      http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookieValue, _ := userService.GetCookieValueByUserID(tt.user.ID)

			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequestGzip())
			router.Use(middleware.RequiredAuthMiddleware(userService))
			router.GET("/api/user/withdrawals", GetWithdrawalsHandler(balanceTransactionService))

			request := httptest.NewRequest("GET", "/api/user/withdrawals", nil)

			if !tt.withoutCookie {
				request.AddCookie(&http.Cookie{
					Name:  "auth",
					Value: cookieValue,
				})
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
