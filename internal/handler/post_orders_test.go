package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

// createTestUser создаёт тестового пользователя в БД
func createTestOrder(t *testing.T, srv service.OrderService, userID int, orderNumber string) *model.Order {
	t.Helper()

	// Создаём заказ
	order, _, err := srv.CreateOrder(context.Background(), userID, orderNumber)
	require.NoError(t, err)

	return order
}

// TestPostOrdersHandlerIntegration тестирование добавления заказа
func TestPostOrdersHandlerIntegration(t *testing.T) {
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
	orderService := service.NewOrderService(cfg, repo)

	// Создаём тестового пользователя
	user := createTestUser(t, userService, "test_user", "test_pass")

	// Создаём тестовый заказ
	order := createTestOrder(t, orderService, user.ID, "5062 8212 3456 7892")

	tests := []struct {
		name          string
		number        string
		withoutCookie bool
		wantCode      int
	}{
		{
			name:     "successful add order",
			number:   "5580 4733 7202 4733",
			wantCode: http.StatusAccepted,
		},
		{
			name:     "duplicate order",
			number:   order.Number,
			wantCode: http.StatusOK,
		},
		{
			name:     "incorrect order number",
			number:   "5062 8217 3456 7892",
			wantCode: http.StatusUnprocessableEntity,
		},
		{
			name:          "without cookie",
			number:        "5580 4733 7202 4733",
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
			router.POST("/api/user/orders", PostOrdersHandler(orderService))

			request := httptest.NewRequest("POST", "/api/user/orders", strings.NewReader(tt.number))
			request.Header.Set("Content-Type", "text/plain")

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
