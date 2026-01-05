package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Password  string    `json:"password"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
