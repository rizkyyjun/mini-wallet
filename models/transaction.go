package models

import (
	"time"
)

type Transaction struct {
	ID          string    `db:"id" json:"id"`
    WalletID    string    `db:"wallet_id" json:"wallet_id"`
    Type        string    `db:"type" json:"type"` // 'deposit' or 'withdrawal'
    Status      string    `db:"status" json:"status"`
    Amount      int64     `db:"amount" json:"amount"`
    ReferenceID string    `db:"reference_id" json:"reference_id"`
    TransactedAt   time.Time `db:"created_at" json:"transacted_at"`
}

type TransactionDTO struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	TransactedAt time.Time `json:"transacted_at"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	ReferenceID string    `json:"reference_id"`
}