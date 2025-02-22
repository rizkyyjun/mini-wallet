package repositories

import (
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
}