# Mini Wallet Exercise

This is a Mini Wallet implementation in Go using PostgreSQL and Redis. The project provides wallet functionalities such as balance retrieval, deposit, withdrawal, and transaction history.

## Prerequisites

Ensure you have the following installed:

- [Go](https://go.dev/dl/) (version 1.20 or later)
- [PostgreSQL](https://www.postgresql.org/download/)
- [Redis](https://redis.io/docs/getting-started/)
- [Git](https://git-scm.com/downloads)

## Installation

### 1. Clone the repository

```sh
git clone https://github.com/rizkyyjun/mini-wallet.git
cd mini-wallet
```

### 2. Set up environment variables

Copy the example environment file and configure it according to your setup:

```sh
cp .env.example .env
```

Modify the `.env` file with your database and Redis credentials.

#### `.env.example`

```
DATABASE_URL=postgresql://<USER>:<PASSWORD>@<HOST>:<PORT>/<DBNAME>?sslmode=require
REDIS_URL=rediss://<USER>:<PASSWORD>@<HOST>:<PORT>
```

### 3. Start PostgreSQL and Redis

Ensure PostgreSQL and Redis are running.

### 4. Create Database Tables

Run the following SQL commands in PostgreSQL to create the necessary tables:

```sql
CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    owned_by UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    enabled_at TIMESTAMP,
    disabled_at TIMESTAMP,
    balance BIGINT NOT NULL
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL, 
    amount BIGINT NOT NULL,
    reference_id UUID UNIQUE NOT NULL,
    transacted_at TIMESTAMP NOT NULL
);

CREATE TABLE customer_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_xid UUID UNIQUE NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### 5. Install dependencies

```sh
go mod tidy
```

## Running the Application

### 1. Start the server

```sh
go run main.go
```

By default, the server should run on `http://localhost:8080`.

### 2. Test the API

Use tools like `curl` or Postman to test endpoints. Example:

```sh
curl -X POST http://localhost:8080/init
```

## Troubleshooting

- Ensure PostgreSQL and Redis are running and accessible.
- Verify that `.env` contains the correct database credentials.
- Run `go mod tidy` if dependencies are missing.

## License

This project is licensed under the MIT License.