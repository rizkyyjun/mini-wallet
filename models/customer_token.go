package models

import "time"

type CustomerToken struct {
	ID			string		`db:"id"`
	CustomerXID	string		`db:"customer_xid"`
	Token		string		`db:"token"`
	CreatedAt	time.Time	`db:"created_at"`
}