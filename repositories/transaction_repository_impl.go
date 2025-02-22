package repositories

import (
	"database/sql"
	"mini-wallet/models"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(transaction *models.Transaction) error {
	query := `INSERT INTO transactions (id, wallet_id, type, status, amount, reference_id, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, transaction.ID, transaction.WalletID, transaction.Type, transaction.Status, transaction.Amount, transaction.ReferenceID, transaction.CreatedAt)
	return err
}

func (r *transactionRepository) GetTransactionByReferenceID(referenceID string) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `SELECT id, wallet_id, type, status, amount, reference_id, created_at FROM transactions WHERE reference_id = $1`
	err := r.db.QueryRow(query, referenceID).Scan(&transaction.ID, &transaction.WalletID, &transaction.Type, &transaction.Status, &transaction.Amount, &transaction.ReferenceID, &transaction.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetTransactionsByWalletID(walletID string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := `SELECT id, wallet_id, type, status, amount, reference_id, created_at FROM transactions WHERE wallet_id = $1`
	rows, err := r.db.Query(query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(&transaction.ID, &transaction.WalletID, &transaction.Type, &transaction.Status, &transaction.Amount, &transaction.ReferenceID, &transaction, transaction.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
