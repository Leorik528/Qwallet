package httpserver

import (
	"encoding/json"
	"time"
)

// TransactionDTO - для внешнего API:
type TransactionDTO struct {
	FromAddress string  `json:"from"`
	ToAddress   string  `json:"to"`
	Amount      float64 `json:"amount"`
}

type ErrorDTO struct {
	Message string

	Time time.Time
}

func (e ErrorDTO) ToString() string {
	b, err := json.MarshalIndent(e, "", "    ")
	if err != nil {
		panic(err)
	}

	return string(b)
}
