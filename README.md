# Pismo Microservices Task

This repository contains a complete microservices architecture implementation for a financial services platform using PostgreSQL as the unified database.

## Architecture Overview

The system consists of three main microservices communicating via gRPC:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gateway       │    │  Account        │    │  Transaction    │
│   Service       │    │  Manager        │    │  Manager        │
│   (Port 8083)   │    │  (Port 8081)    │    │  (Port 8082)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │ gRPC                  │                       │
         ├───────────────────────┼───────────────────────┤
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP/REST     │    │   PostgreSQL    │    │   PostgreSQL    │
│   API           │    │   Database      │    │   Database      │
│                 │    │   (Unified)     │    │   (Unified)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Services

### Gateway Service (Port 8083)
- HTTP REST API for external clients
- Routes requests to appropriate microservices
- CORS support for web applications
- Error handling and response formatting

### Account Manager Service (Port 8081)
- Manages account lifecycle (CRUD operations)
- Handles account balance operations
- Uses PostgreSQL for data persistence
- gRPC service interface

### Transaction Manager Service (Port 8082)
- Processes financial transactions
- Supports multiple operation types:
  - `PAYMENT` - Credit to account
  - `CASH_PURCHASE` - Debit from account
  - `INSTALLMENT_PURCHASE` - Debit from account
  - `WITHDRAWAL` - Debit from account
- Maintains transaction history
- Uses PostgreSQL for data persistence
- gRPC service interface

## Database Schema

### Accounts Table
```sql
CREATE TABLE accounts (
    id VARCHAR(36) PRIMARY KEY,
    document_number VARCHAR(20) NOT NULL UNIQUE,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('CHECKING', 'SAVINGS', 'CREDIT')),
    balance DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(36) NOT NULL,
    operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN ('CASH_PURCHASE', 'INSTALLMENT_PURCHASE', 'WITHDRAWAL', 'PAYMENT')),
    amount DECIMAL(15,2) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);
```

## API Endpoints

### Account Management
- `POST /accounts` - Create account
- `GET /accounts/{id}` - Get account details
- `GET /accounts/{id}/balance` - Get account balance

### Transaction Management
- `POST /transactions` - Create transaction
- `GET /transactions/{id}` - Get transaction details
- `GET /accounts/{account_id}/transactions` - Get transaction history
- `POST /payments` - Process payment

### System
- `GET /health` - Health check

## Development Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- Protocol Buffers compiler (`protoc`)

### Database Setup
1. Start PostgreSQL:
   ```bash
   brew services start postgresql@14
   ```

2. Create database and user:
   ```bash
   createdb pismo
   psql -d pismo -c "CREATE USER pismo WITH PASSWORD 'pismo123';"
   psql -d pismo -c "GRANT ALL PRIVILEGES ON DATABASE pismo TO pismo;"
   ```

3. Initialize database schema:
   ```bash
   psql -d pismo -f services/database/init.sql
   ```

### Running the Services

1. **Start Account Manager:**
   ```bash
   cd services/account-manager
   go build -o account-manager .
   ./account-manager
   ```

2. **Start Transaction Manager:**
   ```bash
   cd services/transaction-manager
   go build -o transaction-manager .
   ./transaction-manager
   ```

3. **Start Gateway:**
   ```bash
   cd services/gateway
   go build -o gateway .
   PORT=8083 ./gateway
   ```

### Testing

Run the comprehensive test suite:
```bash
./test-unified-system.sh
```

Run individual service tests:
```bash
./test-account-service.sh
./test-transaction-service.sh
```

## Project Structure

```
├── api/
│   └── proto/                    # Protocol buffer definitions
│       ├── account/              # Account service generated code
│       └── transaction/          # Transaction service generated code
├── services/
│   ├── account-manager/          # Account management service
│   ├── transaction-manager/      # Transaction processing service
│   ├── gateway/                  # HTTP API gateway
│   └── database/                 # Database initialization scripts
├── tests/                        # Test files
├── test-*.sh                     # Test scripts
└── README.md                     # This file
```

## Key Features

- **Unified Database**: Single PostgreSQL instance for all services
- **ACID Transactions**: Strong consistency across operations
- **gRPC Communication**: High-performance inter-service communication
- **REST API**: Easy integration for external clients
- **Data Validation**: Database-level constraints and validation
- **Error Handling**: Comprehensive error responses
- **Health Checks**: Service monitoring capabilities

## Architecture Benefits

- **Simplified Deployment**: Single database system
- **Data Integrity**: ACID transactions across services
- **Consistent Data Model**: Unified schema across services
- **Better Performance**: No cross-database queries
- **Easier Maintenance**: Single database to manage
- **Strong Consistency**: Immediate data consistency