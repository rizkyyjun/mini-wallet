package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"

	"mini-wallet/models"
	"mini-wallet/repositories"

	"github.com/gin-gonic/gin"
)

type InitHandler struct {
	walletRepo              repositories.WalletRepository
	customerTokenRepository repositories.CustomerTokenRepository
}

func NewInitHandler(walletRepo repositories.WalletRepository, customerTokenRepo repositories.CustomerTokenRepository) *InitHandler {
	return &InitHandler{
		walletRepo:              walletRepo,
		customerTokenRepository: customerTokenRepo,
	}
}

func (h *InitHandler) Init(c *gin.Context) {
	var request struct {
		CustomerXID string `form:"customer_xid" binding:"required"`
	}
	if err := c.ShouldBind(&request); err != nil {
		// customer_xid is missing
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": gin.H{
					"customer_xid": []string{"Missing data for required field."},
				},
			},
		})
		return
	}

	// Check if customer already exists
	exists, err := h.customerTokenRepository.CustomerExists(request.CustomerXID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to check customer existence",
		})
		return
	}

	var token string
	if !exists {
		// Generate token
		token = generateToken(request.CustomerXID)

		// Save token in the db
		if err := h.customerTokenRepository.CreateToken(request.CustomerXID, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to create token",
			})
			return
		}

		// Create a new wallet for the customer
		wallet := models.Wallet{
			ID:         request.CustomerXID,
			OwnedBy:    request.CustomerXID,
			Status:     "disabled",
			EnabledAt:  time.Time{},
			DisabledAt: time.Time{},
			Balance:    0,
		}
		if err := h.walletRepo.CreateWallet(&wallet); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to create wallet",
			})
			return
		}
	} else {
		// retrieve token from db
		token, err = h.customerTokenRepository.GetToken(request.CustomerXID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to retrieve token",
			})
			return
		}
	}

	// return the token in the response
	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": gin.H{
			"token": token,
		},
	})
}

func generateToken(customerXID string) string {
	hash := sha1.New()
	hash.Write([]byte(customerXID + time.Now().String()))
	return hex.EncodeToString(hash.Sum(nil))
}
