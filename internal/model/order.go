package model

import (
	"encoding/json"
	"time"
)

type Order struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Number    string     `json:"number"`
	Status    string     `json:"status"`
	Accrual   float32    `json:"accrual"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type OrderResponse struct {
	ID        int        `json:"id"`
	Number    string     `json:"number"`
	Status    string     `json:"status"`
	Accrual   float32    `json:"accrual"`
	CreatedAt *time.Time `json:"created_at"`
}

// MarshalJSON в данных JSON даты должны выводиться в формате RFC3339
func (o Order) MarshalJSON() ([]byte, error) {
	// чтобы избежать рекурсии при json.Marshal, объявляем новый тип
	type OrderAlias Order

	aliasValue := struct {
		OrderAlias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		// встраиваем значение всех полей изначального объекта (embedding)
		OrderAlias: OrderAlias(o),
		// задаём значение для переопределённого поля
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
	}

	return json.Marshal(aliasValue) // вызываем стандартный Marshal
}

// ToResponse конвертирует Order в OrderResponse
func (o Order) ToResponse() OrderResponse {
	return OrderResponse{
		ID:        o.ID,
		Number:    o.Number,
		Status:    o.Status,
		Accrual:   o.Accrual,
		CreatedAt: o.CreatedAt,
	}
}
