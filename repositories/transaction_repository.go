package repositories

import "mini-wallet/models"

type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionByReferenceID(referenceID string) (*models.Transaction, error)
	GetTransactionsByWalletID(walletID string) ([]models.Transaction, error)
}
