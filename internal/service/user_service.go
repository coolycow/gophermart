package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/coolycow/gophermart/internal/config"
	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// UserService Сервис для работы в Handler
type UserService interface {
	GetUserByID(ctx context.Context, userID int) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	GetUserByLoginAndPassword(ctx context.Context, login string, password string) (*model.User, error)

	CreateUser(ctx context.Context, user model.UserRegister) (*model.User, error)
	DeleteUser(ctx context.Context, userID int) error

	GetUserIDFromCookie(cookie *http.Cookie) (int, error)
	GetCookieValueByUser(user *model.User) (string, error)
	GetCookieValueByUserID(userID int) (string, error)
}

// Реализация сервисного слоя
type userService struct {
	repo      repository.Repository
	cfg       *config.Config
	validator *validator.Validate
}

// NewUserService инициализация сервиса
func NewUserService(cfg *config.Config, repo repository.Repository) UserService {
	return &userService{
		repo:      repo,
		cfg:       cfg,
		validator: validator.New(),
	}
}

// generateRandom генерация случайных байт
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// GetUserByID возвращает пользователя по его ID
func (s *userService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUserByLogin возвращает пользователя по его логину
func (s *userService) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

// GetUserByLoginAndPassword возвращает пользователя по его логину и паролю
func (s *userService) GetUserByLoginAndPassword(ctx context.Context, login string, password string) (*model.User, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, httpError.CustomError{
			Message:    "user not found",
			StatusCode: http.StatusUnauthorized,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusUnauthorized,
		}
	}

	return user, nil
}

// CreateUser создаёт нового пользователя
func (s *userService) CreateUser(ctx context.Context, user model.UserRegister) (*model.User, error) {
	// Валидация логина
	if err := s.validator.Var(user.Login, fmt.Sprintf("required,min=%d,max=%d", s.cfg.MinLoginLength, s.cfg.MaxLoginLength)); err != nil {
		return nil, httpError.CustomError{
			Message:    fmt.Sprintf("login validation failed: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Валидация пароля
	if err := s.validator.Var(user.Password, fmt.Sprintf("required,min=%d,max=%d", s.cfg.MinPasswordLength, s.cfg.MaxPasswordLength)); err != nil {
		return nil, httpError.CustomError{
			Message:    fmt.Sprintf("password validation failed: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	existingUser, err := s.repo.GetUserByLogin(ctx, user.Login)

	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if existingUser != nil {
		return nil, httpError.CustomError{
			Message:    "user already exists",
			StatusCode: http.StatusConflict,
		}
	}

	hashedPassword, err := hashPassword(user.Password)

	if err != nil {
		return nil, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return s.repo.CreateUser(ctx, user.Login, hashedPassword)
}

// DeleteUser удаляет пользователя
func (s *userService) DeleteUser(ctx context.Context, userID int) error {
	return s.repo.DeleteUser(ctx, userID)
}

// GetUserIDFromCookie достаёт UserID из переданной куки
func (s *userService) GetUserIDFromCookie(cookie *http.Cookie) (int, error) {
	cookieValue := cookie.Value

	// Декодируем hex
	data, err := hex.DecodeString(cookieValue)
	if err != nil {
		return 0, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if len(data) == 0 {
		return 0, httpError.CustomError{
			Message:    "invalid cookie",
			StatusCode: http.StatusInternalServerError,
		}
	}

	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return 0, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return 0, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// создаём вектор инициализации
	nonceSize := aesGCM.NonceSize()

	// Разделяем nonce и зашифрованные данные
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// расшифровываем
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	userID, err := strconv.Atoi(string(decrypted))

	if err != nil {
		return 0, httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return userID, nil
}

// GetCookieValueByUser возвращает значение куки для указанного user
func (s *userService) GetCookieValueByUser(user *model.User) (string, error) {
	return s.GetCookieValueByUserID(user.ID)
}

// GetCookieValueByUserID возвращает значение куки для указанного userID
func (s *userService) GetCookieValueByUserID(userID int) (string, error) {
	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return "", httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	dst := aesGCM.Seal(nil, nonce, []byte(strconv.Itoa(userID)), nil)

	// Сохраняем nonce вместе с зашифрованными данными
	result := append(nonce, dst...)

	return hex.EncodeToString(result), nil
}

// hashPassword хэширует пароль и возвращает строку
func hashPassword(password string) (string, error) {
	// Чем больше cost тем больше итераций хеширования.
	// При cost=14 выполняется 16384 итераций, а при cost=10 - всего 1024 итераций.
	// При cost=14 время входа 781.1083ms, при cost=10 - 49.4006ms.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
