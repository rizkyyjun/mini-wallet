package handlers

import (
	"net/http"
	"time"

	"mini-wallet/models"
	"mini-wallet/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	walletRepo        repositories.WalletRepository
	transactionRepo   repositories.TransactionRepository
	customerTokenRepo repositories.CustomerTokenRepository
}

func NewWalletHandler(walletRepo repositories.WalletRepository, transactionRepo repositories.TransactionRepository, customerTokenRepo repositories.CustomerTokenRepository) *WalletHandler {
	return &WalletHandler{
		walletRepo:        walletRepo,
		transactionRepo:   transactionRepo,
		customerTokenRepo: customerTokenRepo,
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

func (h *WalletHandler) Deposit(c *gin.Context) {
	var request struct {
		Amount      int64  `json:"amount"`
		ReferenceID string `json:"reference_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	// Check if the referenceId already exists
	_, err := h.transactionRepo.GetTransactionByReferenceID(request.ReferenceID)
	if err != nil {
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
		ID:          uuid.New().String(),
		WalletID:    "wallet-id",
		Type:        "deposiot",
		Status:      "success",
		Amount:      request.Amount,
		ReferenceID: request.ReferenceID,
		CreatedAt:   time.Now().UTC(),
	}

	if err := h.transactionRepo.CreateTransaction(&transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"deposit": transaction,
		},
	})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	var request struct {
		Amount      int64  `json:"amount"`
		ReferenceID string `json:"reference_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	// Check if the referenceId already exists
	_, err := h.transactionRepo.GetTransactionByReferenceID(request.ReferenceID)
	if err != nil {
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
		ID:          uuid.New().String(),
		WalletID:    "wallet-id",
		Type:        "withdraw",
		Status:      "success",
		Amount:      request.Amount,
		ReferenceID: request.ReferenceID,
		CreatedAt:   time.Now().UTC(),
	}

	if err := h.transactionRepo.CreateTransaction(&transaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"withdrawal": transaction,
		},
	})
}

func (h *WalletHandler) DisableWallet(c *gin.Context) {
	var wallet models.Wallet
	if err := c.ShouldBindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	wallet.Status = "disabled"
	wallet.DisabledAt = time.Now().UTC()

	if err := h.walletRepo.DisableWallet(&wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"wallet": wallet,
		},
	})
}
