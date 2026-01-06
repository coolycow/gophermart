package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/coolycow/gophermart/internal/config"
	httpError "github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/model"
	"github.com/coolycow/gophermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// UserService Сервис для работы в Handler
type UserService interface {
	GetUserByID(ctx context.Context, userID int) (model.User, error)
	GetUserByLogin(ctx context.Context, login string) (model.User, error)
	GetUserByLoginAndPassword(ctx context.Context, login string, password string) (model.User, error)
	CreateUser(ctx context.Context, user model.UserRegister) (model.User, error)
	DeleteUser(ctx context.Context, userID int) error

	GetUserIDFromCookie(cookie *http.Cookie) (int, error)
	GetCookieValueByUser(user model.User) (string, error)
	GetCookieValueByUserID(userID int) (string, error)
}

// Реализация сервисного слоя
type userService struct {
	repo repository.Repository
	cfg  *config.Config
}

// NewUserService инициализация сервиса
func NewUserService(cfg *config.Config, repo repository.Repository) UserService {
	return &userService{
		repo: repo,
		cfg:  cfg,
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
func (s *userService) GetUserByID(ctx context.Context, userID int) (model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUserByLogin возвращает пользователя по его логину
func (s *userService) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

// GetUserByLoginAndPassword возвращает пользователя по его логину и паролю
func (s *userService) GetUserByLoginAndPassword(ctx context.Context, login string, password string) (model.User, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)

	if err != nil {
		return model.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

// CreateUser создаёт нового пользователя
func (s *userService) CreateUser(ctx context.Context, user model.UserRegister) (model.User, error) {
	existingUser, err := s.repo.GetUserByLogin(ctx, user.Login)

	if err != nil {
		return model.User{}, err
	}

	if existingUser.ID != 0 {
		return model.User{}, httpError.CustomError{
			Message:    "user already exists",
			StatusCode: http.StatusConflict, // 409
		}
	}

	hashPassword, err := HashPassword(user.Password)

	if err != nil {
		return model.User{}, err
	}

	return s.repo.CreateUser(ctx, user.Login, hashPassword)
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
		return 0, fmt.Errorf("failed to decode hex cookie value: %w", err)
	}

	if len(data) == 0 {
		return 0, errors.New("invalid cookie")
	}

	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return 0, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return 0, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonceSize := aesGCM.NonceSize()

	// Разделяем nonce и зашифрованные данные
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// расшифровываем
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	userID := int(binary.LittleEndian.Uint64(decrypted))

	return userID, nil
}

// GetCookieValueByUser возвращает значение куки для указанного user
func (s *userService) GetCookieValueByUser(user model.User) (string, error) {
	return s.GetCookieValueByUserID(user.ID)
}

// GetCookieValueByUserID возвращает значение куки для указанного userID
func (s *userService) GetCookieValueByUserID(userID int) (string, error) {
	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	dst := aesGCM.Seal(nil, nonce, []byte(strconv.Itoa(userID)), nil)

	// Сохраняем nonce вместе с зашифрованными данными
	result := append(nonce, dst...)

	return hex.EncodeToString(result), nil
}

// HashPassword хэширует пароль и возвращает строку
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
