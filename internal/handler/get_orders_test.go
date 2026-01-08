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

// TestGetOrdersHandlerIntegration тестирование отдачи списка всех заказов
func TestGetOrdersHandlerIntegration(t *testing.T) {
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
	user1 := createTestUser(t, userService, "test_user_1", "test_pass_1")
	user2 := createTestUser(t, userService, "test_user_2", "test_pass_2")

	// Создаём тестовые заказы
	_ = createTestOrder(t, orderService, user1.ID, "0018")
	_ = createTestOrder(t, orderService, user1.ID, "5062 8212 3456 7892")

	tests := []struct {
		name          string
		user          *model.User
		withoutCookie bool
		wantCode      int
	}{
		{
			name:     "successful orders",
			user:     user1,
			wantCode: http.StatusOK,
		},
		{
			name:     "no content",
			user:     user2,
			wantCode: http.StatusNoContent,
		},
		{
			name:          "without cookie",
			user:          user2,
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
			router.GET("/api/user/orders", GetOrdersHandler(orderService))

			request := httptest.NewRequest("GET", "/api/user/orders", nil)
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
