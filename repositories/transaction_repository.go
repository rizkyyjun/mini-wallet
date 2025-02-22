package repositories

import (
	"database/sql"
	"mini-wallet/models"
)

type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	GetTransactionByReferenceID(referenceID string) (*models.Transaction, error)
	GetTransactionsByWalletID(walletID string) ([]models.Transaction, error)
	CreateTransactionWithTx(tx *sql.Tx, transaction *models.Transaction) error
}
