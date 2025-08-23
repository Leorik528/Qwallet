package qwalletstore

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Qwallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}

type QwalletList struct {
	Qwallets map[string]Qwallet
	mtx      sync.RWMutex
}

type TrasactionList struct {
	Transactions map[string]Transaction
	mtx          sync.RWMutex
}

// Transaction - для внутренней бизнес-логики и БД
type Transaction struct {
	ID          int64 // Для
	FromAddress string
	ToAddress   string
	Amount      float64
	CreatedAt   time.Time // Для TIMESTAMP
}

func GenerateRandomAddress() string {

	b := make([]byte, 16) // генерируем 16 случайных байт
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
