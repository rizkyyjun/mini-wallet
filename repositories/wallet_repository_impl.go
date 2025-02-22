package repositories

import (
	"database/sql"
	"mini-wallet/models"
	"time"
)

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{db:db}
}

func (r *walletRepository) GetWalletByCustomerXID(customerXID string) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT id, owned_by, status, enabled_at, balance FROM wallets WHERE owned_by = $1`
	err := r.db.QueryRow(query, customerXID).Scan(&wallet.ID, &wallet.OwnedBy, &wallet.Status, &wallet.EnabledAt, &wallet.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No wallet found
		}
		return nil, err
	}
	return &wallet, err
}

func (r *walletRepository) CreateWallet(wallet *models.Wallet) error {
	query := `INSERT INTO wallets (id, owned_by, status, enabled_at, disabled_at, balance)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, wallet.ID, wallet.OwnedBy, wallet.Status, wallet.EnabledAt, wallet.DisabledAt, wallet.Balance)
	return err
}

func (r *walletRepository) GetWalletByID(id string) (*models.Wallet, error) {
	var wallet models.Wallet
	query := `SELECT id, owned_by, status, enabled_at, disabled_at, balance FROM wallets
	WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&wallet.ID, &wallet.OwnedBy, &wallet.Status, &wallet.EnabledAt, &wallet.DisabledAt, &wallet.Balance)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) UpdateWalletStatus(walletID string, status string, enabledAt time.Time) error {
	query := `UPDATE wallets SET status = $1, enabled_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, enabledAt, walletID)
	return err
}

func (r *walletRepository) UpdateWallet(wallet *models.Wallet) error {
	query := `UPDATE wallets SET status = $1, enabled_at = $2, disabled_at = $3, balance = $4 WHERE id = $5`
	_, err := r.db.Exec(query, wallet.Status, wallet.EnabledAt, wallet.DisabledAt, wallet.Balance, wallet.ID)
	return err
}

func (r *walletRepository) UpdateWalletBalance(walletID string, newBalance int64) error {
	query := `
	UPDATE wallets
	SET balance = $1
	WHERE id = $2`
	_, err := r.db.Exec(query, newBalance, walletID)
	return err
}

func (r *walletRepository) UpdateWalletBalanceWithTx(tx *sql.Tx, walletID string, balance int64) error {
    query := `UPDATE wallets SET balance = $1 WHERE id = $2`
    _, err := tx.Exec(query, balance, walletID)
    return err
}

func (r *walletRepository) WithTransaction(fn func(tx *sql.Tx) error) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }

    // Execute the provided function within the transaction
    if err := fn(tx); err != nil {
        // Rollback on error
        if rollbackErr := tx.Rollback(); rollbackErr != nil {
            return rollbackErr
        }
        return err
    }

    // Commit the transaction if no errors
    return tx.Commit()
}