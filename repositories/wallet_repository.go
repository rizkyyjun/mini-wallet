package repositories

import (
	"database/sql"
	"mini-wallet/models"
	"time"
)

type WalletRepository interface {
	GetWalletByCustomerXID(customerXID string) (*models.Wallet, error)
	CreateWallet(wallet *models.Wallet) error
	GetWalletByID(id string) (*models.Wallet, error)
	UpdateWallet(wallet *models.Wallet) error
	UpdateWalletStatus(walletID string, status string, enabledAt time.Time) error
	UpdateWalletBalance(walletID string, newBalance int64) error
	WithTransaction(fn func(tx *sql.Tx) error) error
	UpdateWalletBalanceWithTx(tx *sql.Tx, walletID string, balance int64) error
}