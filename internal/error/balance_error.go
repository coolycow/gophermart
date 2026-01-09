package error

import "net/http"

type InsufficientBalanceError struct {
	Message string
}

func (e *InsufficientBalanceError) Error() string {
	return e.Message
}

func (e *InsufficientBalanceError) StatusCode() int {
	return http.StatusPaymentRequired // 402
}
