package models

import (
	"time"
)

type Wallet struct {
	ID         string    `db:"id" json:"id"`
    OwnedBy    string    `db:"owned_by" json:"owned_by"`
    Status     string    `db:"status" json:"status"`
    EnabledAt  time.Time `db:"enabled_at" json:"enabled_at"`
    DisabledAt time.Time `db:"disabled_at" json:"disabled_at"`
    Balance    int64     `db:"balance" json:"balance"`
}