package model

/**
Про статусы:
    REGISTERED — заказ зарегистрирован, но вознаграждение не рассчитано;
    INVALID — заказ не принят к расчёту, и вознаграждение не будет начислено;
    PROCESSING — расчёт начисления в процессе;
    PROCESSED — расчёт начисления окончен;

	Статусы INVALID и PROCESSED являются окончательными.
*/

type Accrual struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual,omitempty"`
}
