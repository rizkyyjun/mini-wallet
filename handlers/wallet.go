package handlers

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"mini-wallet/models"
	"mini-wallet/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletRepo        repositories.WalletRepository
	transactionRepo   repositories.TransactionRepository
	customerTokenRepo repositories.CustomerTokenRepository
	redisClient       *redis.Client
}

func NewWalletHandler(walletRepo repositories.WalletRepository, transactionRepo repositories.TransactionRepository, customerTokenRepo repositories.CustomerTokenRepository, redisClient *redis.Client) *WalletHandler {
	return &WalletHandler{
		walletRepo:        walletRepo,
		transactionRepo:   transactionRepo,
		customerTokenRepo: customerTokenRepo,
		redisClient:       redisClient,
	}
}

func (h *WalletHandler) EnableWallet(c *gin.Context) {
	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Authorization token is required",
			},
		})
		return
	}

	// Extract token without "Token" prefix
	token = token[6:]

	// Get customer_xid by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Invalid token",
			},
		})
	}

	// Check if wallet exists
	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		// Create wallet if not exists
		wallet = &models.Wallet{
			ID:        uuid.New().String(),
			OwnedBy:   customerXID,
			Status:    "enabled",
			EnabledAt: time.Now().UTC(),
			Balance:   0,
		}

		if err := h.walletRepo.CreateWallet(wallet); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
	} else {
		// Fail if wallet is already enabled
		if wallet.Status == "enabled" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "fail",
				"data": gin.H{
					"error": "Already enabled",
				},
			})
			return
		}

		// Enable wallet
		wallet.Status = "enabled"
		wallet.EnabledAt = time.Now().UTC()
		if err := h.walletRepo.UpdateWalletStatus(wallet.ID, "enabled", wallet.EnabledAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
	}

	// Return response
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"wallet": gin.H{
				"id":         wallet.ID,
				"owned_by":   wallet.OwnedBy,
				"status":     wallet.Status,
				"enabled_at": wallet.EnabledAt,
				"balance":    wallet.Balance,
			},
		},
	})

}

func (h *WalletHandler) ViewWalletBalance(c *gin.Context) {
	ctx := context.Background()

	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Authorization token is required",
			},
		})
		return
	}
	token = token[6:]

	// Get customerXID by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Invalid token",
			},
		})
		return
	}

	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet not found",
			},
		})
		return
	}

	// wallet is disabled
	if wallet.Status == "disabled" {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet disabled",
			},
		})
		return
	}

	// Calculate balance from transactions for consistency
	transactions, err := h.transactionRepo.GetTransactionsByWalletID(wallet.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch transactions"})
		return
	}
	var balance int64
	for _, t := range transactions {
		if t.Type == "deposit" {
			balance += t.Amount
		} else if t.Type == "withdrawal" {
			balance -= t.Amount
		}
	}

	// Update Redis cache
	cacheKey := "wallet_balance:" + customerXID
	if err := h.redisClient.Set(ctx, cacheKey, wallet.Balance, 15*time.Second).Err(); err != nil {
		log.Println("Failed to update Redis cache:", err)
	}

	// Fetch from db
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"wallet": gin.H{
				"id":         wallet.ID,
				"owned_by":   wallet.OwnedBy,
				"status":     wallet.Status,
				"enabled_at": wallet.EnabledAt,
				"balance":    wallet.Balance,
			},
		},
	})
}

func (h *WalletHandler) ViewWalletTransactions(c *gin.Context) {
	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Authorization token is required",
			},
		})
		return
	}
	token = token[6:]

	// Get customerXID by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Invalid token",
			},
		})
		return
	}

	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet not found",
			},
		})
		return
	}

	// wallet is disabled
	if wallet.Status == "disabled" {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet disabled",
			},
		})
		return
	}

	// Get list transaction from the wallet
	transactions, err := h.transactionRepo.GetTransactionsByWalletID(wallet.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve transactions",
		})
		return
	}

	var transactionsDTO []models.TransactionDTO
	for _, transaction := range transactions {
		transactionsDTO = append(transactionsDTO, models.TransactionDTO{
			ID:           transaction.ID,
			Status:       transaction.Status,
			TransactedAt: transaction.TransactedAt,
			Type:         transaction.Type,
			Amount:       transaction.Amount,
			ReferenceID:  transaction.ReferenceID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"transactions": transactionsDTO,
		},
	})

}

func (h *WalletHandler) Deposit(c *gin.Context) {
	ctx := context.Background()

	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Authorization token is required",
			},
		})
		return
	}
	token = token[6:]

	// Get customerXID by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Invalid token",
			},
		})
		return
	}

	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet not found",
			},
		})
		return
	}

	// wallet is disabled
	if wallet.Status == "disabled" {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet disabled",
			},
		})
		return
	}

	// parse form data
	amountStr := c.PostForm("amount")
	referenceID := c.PostForm("reference_id")

	if amountStr == "" || referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "data": gin.H{"error": "amount and reference_id are required"}})
		return
	}

	// Convert amount to int64
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "data": gin.H{"error": "invalid amount format"}})
		return
	}

	// Check if the referenceId already exists
	if _, err := h.transactionRepo.GetTransactionByReferenceID(referenceID); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"reference_id": "duplicate reference_id",
			},
		})
		return
	}

	// Create the transaction
	transaction := models.Transaction{
		ID:           uuid.New().String(),
		WalletID:     wallet.ID,
		Type:         "deposit",
		Status:       "success",
		Amount:       amount,
		ReferenceID:  referenceID,
		TransactedAt: time.Now().UTC(),
	}

	// Record the transaction
	err = h.transactionRepo.CreateTransaction(&transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to record transaction"})
		return
	}

	// Defer balance update with random delay up to 5 seconds
	go func(walletID string, customerXID string) {
		delay := time.Duration(rand.Intn(5)) * time.Second
		time.Sleep(delay)

		// Acquire lock to ensure atomic balance update
		lockKey := "lock:wallet:" + customerXID
		lock := h.redisClient.SetNX(ctx, lockKey, "1", 15*time.Second)
		if err := lock.Err(); err != nil || !lock.Val() {
			log.Printf("Failed to acquire lock for balance update of wallet %s", walletID)
			return
		}
		defer h.redisClient.Del(ctx, lockKey)

		// Calculate balance from transactions
		transactions, err := h.transactionRepo.GetTransactionsByWalletID(walletID)
		if err != nil {
			log.Printf("Failed to fetch transactions for wallet %s: %v", walletID, err)
			return
		}
		var newBalance int64
		for _, t := range transactions {
			if t.Type == "deposit" {
				newBalance += t.Amount
			} else if t.Type == "withdrawal" {
				newBalance -= t.Amount
			}
		}

		// Update database balance
		if err := h.walletRepo.UpdateWalletBalance(walletID, newBalance); err != nil {
			log.Printf("Failed to update database balance for wallet %s: %v", walletID, err)
			return
		}

		// Update redis balance
		cacheKey := "wallet_balance:" + customerXID
		if err := h.redisClient.Set(ctx, cacheKey, newBalance, 15*time.Second).Err(); err != nil {
			log.Printf("Failed to update Redis balance for %s: %v", cacheKey, err)
			h.redisClient.Del(ctx, cacheKey)
		}
	}(wallet.ID, customerXID)

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"deposit": gin.H{
				"id":           transaction.ID,
				"deposited_by": customerXID,
				"status":       transaction.Status,
				"deposited_at": transaction.TransactedAt,
				"amount":       transaction.Amount,
				"reference_id": transaction.ReferenceID,
			},
		},
	})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	ctx := context.Background()

	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Authorization token is required",
			},
		})
		return
	}
	token = token[6:]

	// Parse form data
	amountStr := c.PostForm("amount")
	referenceID := c.PostForm("reference_id")
	if amountStr == "" || referenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "data": gin.H{"error": "amount and reference_id are required"}})
		return
	}

	// Get customerXID by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Invalid token",
			},
		})
		return
	}

	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet not found",
			},
		})
		return
	}

	// wallet is disabled
	if wallet.Status == "disabled" {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Wallet disabled",
			},
		})
		return
	}

	// convert amount to int64
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "data": gin.H{"error": "invalid amount format"}})
		return
	}

	// Check if the referenceId already exists
	if _, err := h.transactionRepo.GetTransactionByReferenceID(referenceID); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"reference_id": "duplicate reference_id",
			},
		})
		return
	}

	// Ensure sufficient balance
	if wallet.Balance < amount {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "data": gin.H{"error": "Insufficient balance"}})
		return
	}

	// Create the transaction
	transaction := models.Transaction{
		ID:           uuid.New().String(),
		WalletID:     wallet.ID,
		Type:         "withdrawal",
		Status:       "success",
		Amount:       amount,
		ReferenceID:  referenceID,
		TransactedAt: time.Now().UTC(),
	}

	// Record the transaction
	err = h.transactionRepo.CreateTransaction(&transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to record transaction"})
		return
	}

	// Defer balance update with a random delay up to 5 seconds
	go func(walletID string, customerXID string) {
		delay := time.Duration(rand.Intn(5)) * time.Second // Random delay between 0-5 seconds
		time.Sleep(delay)

		// Acquire lock to ensure atomic balance update
		lockKey := "lock:wallet:" + customerXID
		lock := h.redisClient.SetNX(ctx, lockKey, "1", 15*time.Second)
		if err := lock.Err(); err != nil || !lock.Val() {
			log.Printf("Failed to acquire lock for balance update of wallet %s", walletID)
			return
		}
		defer h.redisClient.Del(ctx, lockKey)

		// Calculate balance from transactions
		transactions, err := h.transactionRepo.GetTransactionsByWalletID(walletID)
		if err != nil {
			log.Printf("Failed to fetch transactions for wallet %s: %v", walletID, err)
			return
		}
		var newBalance int64
		for _, t := range transactions {
			if t.Type == "deposit" {
				newBalance += t.Amount
			} else if t.Type == "withdrawal" {
				newBalance -= t.Amount
			}
		}

		// Update database balance
		if err := h.walletRepo.UpdateWalletBalance(walletID, newBalance); err != nil {
			log.Printf("Failed to update database balance for wallet %s: %v", walletID, err)
			return
		}

		// Update Redis balance
		cacheKey := "wallet_balance:" + customerXID
		if err := h.redisClient.Set(ctx, cacheKey, newBalance, 15*time.Second).Err(); err != nil {
			log.Printf("Failed to update Redis balance for %s: %v", cacheKey, err)
			h.redisClient.Del(ctx, cacheKey)
		}
	}(wallet.ID, customerXID)

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"withdrawal": gin.H{
				"id":           transaction.ID,
				"withdrawn_by": customerXID,
				"status":       transaction.Status,
				"withdrawn_at": transaction.TransactedAt,
				"amount":       transaction.Amount,
				"reference_id": transaction.ReferenceID,
			},
		},
	})
}

func (h *WalletHandler) DisableWallet(c *gin.Context) {
	// Get Authorization token from header
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "Authorization token is required"},
		})
		return
	}
	token = token[6:]

	// Get customerXID by token
	customerXID, err := h.customerTokenRepo.GetCustomerXIDByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "Invalid token"},
		})
		return
	}

	// Fetch the customer's wallet
	wallet, err := h.walletRepo.GetWalletByCustomerXID(customerXID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "Wallet not found"},
		})
		return
	}

	// Check if wallet is already disabled
	if wallet.Status == "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "Wallet is already disabled"},
		})
		return
	}

	// Read `is_disabled` from form-data
	isDisabled := c.PostForm("is_disabled")
	if isDisabled == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "is_disabled is required"},
		})
		return
	}

	if isDisabled != "true" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data":   gin.H{"error": "Invalid request: is_disabled must be 'true'"},
		})
		return
	}

	// Disable wallet
	disabledAt := time.Now().UTC()
	if err := h.walletRepo.UpdateWalletStatus(wallet.ID, "disabled", disabledAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to disable wallet",
		})
		return
	}

	wallet.Status = "disabled"
	wallet.DisabledAt = disabledAt

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"wallet": gin.H{
				"id":          wallet.ID,
				"owned_by":    customerXID,
				"status":      wallet.Status,
				"disabled_at": wallet.DisabledAt.Format(time.RFC3339),
				"balance":     wallet.Balance,
			},
		},
	})
}
