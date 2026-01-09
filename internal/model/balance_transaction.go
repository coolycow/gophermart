package model

import (
	"encoding/json"
	"math"
	"time"
)

// WithdrawalRequest - модель для запроса на списание баллов
type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

// BalanceResponse - модель для ответа на запрос о состоянии баланса пользователя
type BalanceResponse struct {
	Current   float32 `json:"current"`   // Текущая сумма баллов лояльности
	Withdrawn float32 `json:"withdrawn"` // Сумма баллов, использованных за весь период
}

// BalanceTransaction - базовая модель баланса (как в БД)
type BalanceTransaction struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	OrderNumber string     `json:"number"`
	Amount      float32    `json:"amount"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// ToWithdrawalResponse - конвертация BalanceTransaction в WithdrawalResponse
func (b BalanceTransaction) ToWithdrawalResponse() WithdrawalResponse {
	return WithdrawalResponse{
		Order:       b.OrderNumber,
		Sum:         b.Amount,
		ProcessedAt: b.CreatedAt,
	}
}

// WithdrawalResponse - модель для ответа на запрос о предоставлении списка списания баллов пользователя
type WithdrawalResponse struct {
	Order       string     `json:"order"`
	Sum         float32    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at"`
}

// MarshalJSON в данных JSON даты должны выводиться в формате RFC3339
func (o WithdrawalResponse) MarshalJSON() ([]byte, error) {
	// чтобы избежать рекурсии при json.Marshal, объявляем новый тип
	type WithdrawalResponseAlias WithdrawalResponse

	aliasValue := struct {
		WithdrawalResponseAlias
		Sum         float32 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}{
		// встраиваем значение всех полей изначального объекта (embedding)
		WithdrawalResponseAlias: WithdrawalResponseAlias(o),
		// задаём значение для переопределённого поля
		Sum:         float32(math.Abs(float64(o.Sum))),
		ProcessedAt: o.ProcessedAt.Format(time.RFC3339),
	}

	return json.Marshal(aliasValue) // вызываем стандартный Marshal
}
