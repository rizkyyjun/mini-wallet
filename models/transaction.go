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
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
}