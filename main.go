package main

import (
    "database/sql"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
    "mini-wallet/handlers"
    "mini-wallet/repositories"
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

    // Initialize repositories
    walletRepo := repositories.NewWalletRepository(db)
    transactionRepo := repositories.NewTransactionRepository(db)
	customerTokenRepo := repositories.NewCustomerTokenRepository(db)

    // Initialize handlers
    walletHandler := handlers.NewWalletHandler(walletRepo, transactionRepo, customerTokenRepo)
	initHandler := handlers.NewInitHandler(walletRepo, customerTokenRepo)

    // Initialize the Gin router
    router := gin.Default()

    // Define API endpoints
	router.POST("/api/v1/init", initHandler.Init)
	router.POST("/api/v1/wallet", walletHandler.EnableWallet)
    router.POST("/api/v1/wallet/deposits", walletHandler.Deposit)
    router.POST("/api/v1/wallet/withdrawals", walletHandler.Withdraw)
    router.PATCH("/api/v1/wallet", walletHandler.DisableWallet)

    // Start the server
    log.Println("Starting server on :8080...")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal("Failed to start the server:", err)
    }
}