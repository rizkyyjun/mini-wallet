package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"mini-wallet/handlers"
	"mini-wallet/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get the connection string from environment variables
	connectionString := os.Getenv("DATABASE_URL")
	if connectionString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Open a connection to the database
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database:", err)
	}
	log.Println("Successfully connected to the database!")

	// Init Redis connection
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set")
	}

	// Parse and connect to Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Invalid Redis URL:", err)
	}

	redisClient := redis.NewClient(opt)

	// Test redis connection
	ctx := context.Background()
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Connected to Redis:", pong)

	// Initialize repositories
	walletRepo := repositories.NewWalletRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	customerTokenRepo := repositories.NewCustomerTokenRepository(db)

	// Initialize handlers
	walletHandler := handlers.NewWalletHandler(walletRepo, transactionRepo, customerTokenRepo, redisClient)
	initHandler := handlers.NewInitHandler(walletRepo, customerTokenRepo)

	// Initialize the Gin router
	router := gin.Default()

	// Define API endpoints
	router.POST("/api/v1/init", initHandler.Init)
	router.POST("/api/v1/wallet", walletHandler.EnableWallet)
	router.GET("/api/v1/wallet", walletHandler.ViewWalletBalance)
	router.GET("/api/v1/wallet/transacions", walletHandler.ViewWalletTransactions)
	router.POST("/api/v1/wallet/deposits", walletHandler.Deposit)
	router.POST("/api/v1/wallet/withdrawals", walletHandler.Withdraw)
	router.PATCH("/api/v1/wallet", walletHandler.DisableWallet)

	// Start the server
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}
