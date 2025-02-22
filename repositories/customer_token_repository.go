package repositories

import (
	"database/sql"
)

type CustomerTokenRepository interface {
	CreateToken(customerXID, token string) error
	GetCustomerXIDByToken(token string) (string, error)
	CustomerExists(customerXID string) (bool, error)
	GetToken(customerXID string) (string, error)
}

type customerTokenRepository struct {
	db *sql.DB
}

func NewCustomerTokenRepository(db *sql.DB) CustomerTokenRepository {
	return &customerTokenRepository{db: db}
}

func (r *customerTokenRepository) CreateToken(customerXID, token string) error {
	query := `INSERT INTO customer_tokens (customer_xid, token) VALUES ($1, $2)`
	_, err := r.db.Exec(query, customerXID, token)
	return err
}

func (r *customerTokenRepository) GetCustomerXIDByToken(token string) (string, error) {
	var customerXID string
	query := `SELECT customer_xid FROM customer_tokens WHERE token = $1`
	err := r.db.QueryRow(query, token).Scan(&customerXID)
	if err != nil {
		return "", err
	}
	return customerXID, nil
}

func (r *customerTokenRepository) CustomerExists(customerXID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM customer_tokens WHERE customer_xid = $1)`
	err := r.db.QueryRow(query, customerXID).Scan(&exists)
	return exists, err
}

func (r *customerTokenRepository) GetToken(customerXID string) (string, error) {
	var token string
	query := `SELECT token FROM customer_tokens WHERE customer_xid = $1`
	err := r.db.QueryRow(query, customerXID).Scan(&token)
	return token, err
}
