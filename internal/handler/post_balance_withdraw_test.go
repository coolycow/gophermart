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

// TestPostBalanceWithdrawHandlerIntegration тестирование добавления заказа
func TestPostBalanceWithdrawHandlerIntegration(t *testing.T) {
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

	// Создаём тестового пользователя
	user := createTestUser(t, userService, "test_user", "test_pass")

	// Создаём тестовое начисление баллов
	_ = createTestAccrual(t, balanceTransactionService, user.ID, "5062 8212 3456 7892", 10000)

	tests := []struct {
		name          string
		order         string
		sum           float64
		withoutCookie bool
		wantCode      int
	}{
		{
			name:     "successful",
			order:    "5580 4733 7202 4733",
			sum:      999.99,
			wantCode: http.StatusOK,
		},
		{
			name:     "insufficient balance",
			order:    "5580 4733 7202 4733",
			sum:      100,
			wantCode: http.StatusOK,
		},
		{
			name:     "incorrect order number",
			order:    "5062 8217 3456 7892",
			sum:      1000,
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:          "without cookie",
			order:         "5580 4733 7202 4733",
			withoutCookie: true,
			wantCode:      http.StatusUnauthorized,
		},
	}

	cookieValue, _ := userService.GetCookieValueByUserID(user.ID)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.Use(middleware.RequestGzip())
			router.Use(middleware.RequiredAuthMiddleware(userService))
			router.POST("/api/user/balance/withdraw", PostBalanceWithdrawHandler(balanceTransactionService))

			jsonData := fmt.Sprintf(`{"order": "%s", "sum": %f}`, tt.order, tt.sum)
			request := httptest.NewRequest("POST", "/api/user/balance/withdraw", strings.NewReader(jsonData))
			request.Header.Set("Content-Type", "application/json")

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
