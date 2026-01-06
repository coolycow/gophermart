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

func LoginHandler(srv service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Log.Debug("start login handler")

		// Получаем данные пользователя из запроса
		var req model.UserLogin
		dec := json.NewDecoder(c.Request.Body)

		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Находим пользователя по паре логин/пароль
		user, err := srv.GetUserByLoginAndPassword(c.Request.Context(), req.Login, req.Password)

		if err != nil {
			logger.Log.Debug("cannot get user by login", zap.Error(err))
			_ = c.Error(err)
			return
		}

		logger.Log.Debug("get user success", zap.Any("user", user))

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
