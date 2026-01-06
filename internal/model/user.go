package model

import "time"

// User полная модель пользователя
type User struct {
	ID        int        `json:"id"`
	Password  string     `json:"password"`
	Login     string     `json:"login"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// UserResponse модель пользователя для ответа API (без пароля и метки удаления)
type UserResponse struct {
	ID        int        `json:"id"`
	Login     string     `json:"login"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// ToResponse конвертирует User в UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Login:     u.Login,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserLogin данные пользователя для входа
type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserRegister данные пользователя для регистрации
type UserRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
