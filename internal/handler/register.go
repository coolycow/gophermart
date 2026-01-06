package handler

import (
	"encoding/json"
	"net/http"

	"github.com/coolycow/gophermart/internal/logger"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RegisterHandler(srv service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные пользователя из запроса
		var req model.UserRegister
		dec := json.NewDecoder(c.Request.Body)

		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Создаем нового пользователя
		user, err := srv.CreateUser(c.Request.Context(), req)
		if err != nil {
			logger.Log.Debug("cannot create user", zap.Error(err))
			_ = c.Error(err)
			return
		}

		logger.Log.Debug("create user success", zap.Any("user", user))

		// Формируем данные для авторизационной куки
		cookieValue, err := srv.GetCookieValueByUser(user)

		if err != nil {
			logger.Log.Debug("cannot get cookie value", zap.Error(err))
			_ = c.Error(err)
			return
		}

		logger.Log.Debug("create cookie success", zap.Any("cookieValue", cookieValue))

		// Устанавливаем куку
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "auth",
			Value:    cookieValue,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
		})

		c.Status(http.StatusOK)
	}
}
