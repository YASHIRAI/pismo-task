# Pismo Microservices Task

This repository contains a complete microservices architecture implementation for a financial services platform using PostgreSQL as the unified database, following Go best practices for project structure.

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

## Project Structure

```
├── cmd/                          # Main applications
│   ├── account-mgr/              # Account management service
│   │   ├── main.go              # Service entry point with go:generate
│   │   ├── go.mod               # Service dependencies
│   │   └── Dockerfile           # Container configuration
│   ├── transaction-mgr/          # Transaction processing service
│   │   ├── main.go              # Service entry point with go:generate
│   │   ├── go.mod               # Service dependencies
│   │   └── Dockerfile           # Container configuration
│   └── gateway/                  # HTTP API gateway
│       ├── main.go              # Gateway entry point
│       ├── go.mod               # Gateway dependencies
│       └── Dockerfile           # Container configuration
├── internal/                     # Private application code
│   ├── common/                   # Shared utilities
│   │   ├── database.go          # Database connection and pooling
│   │   ├── orm.go               # Database models
│   │   ├── proto_db.go          # Protobuf ↔ Database conversion
│   │   └── go.mod               # Common dependencies
│   ├── account/                  # Account business logic
│   │   ├── account.go           # Account service implementation
│   │   └── go.mod               # Account dependencies
│   ├── transaction/              # Transaction business logic
│   │   ├── transaction.go       # Transaction service implementation
│   │   └── go.mod               # Transaction dependencies
│   └── health/                   # Health check utilities
│       ├── health.go            # Health check implementation
│       └── go.mod               # Health dependencies
├── proto/                        # Protocol buffer definitions
│   ├── account.proto            # Account service definitions
│   ├── transaction.proto        # Transaction service definitions
│   └── health.proto             # Health check definitions
├── scripts/                      # Database initialization scripts
│   └── database/                 # Database setup scripts
│       └── init.sql             # Database schema initialization
├── tests/                        # Test files
├── test-*.sh                     # Test scripts
├── go.mod                        # Root module dependencies
└── README.md                     # This file
```

## Services

### Gateway Service (cmd/gateway)
- **Port**: 8083
- **Purpose**: HTTP REST API for external clients
- **Features**:
  - Routes requests to appropriate microservices
  - CORS support for web applications
  - Error handling and response formatting
  - Health check endpoint

### Account Manager Service (cmd/account-mgr)
- **Port**: 8081
- **Purpose**: Manages account lifecycle and operations
- **Features**:
  - Account CRUD operations
  - Balance management
  - PostgreSQL integration
  - gRPC service interface
  - Auto-generated protobuf code

### Transaction Manager Service (cmd/transaction-mgr)
- **Port**: 8082
- **Purpose**: Processes financial transactions
- **Features**:
  - Multiple operation types:
    - `PAYMENT` - Credit to account
    - `CASH_PURCHASE` - Debit from account
    - `INSTALLMENT_PURCHASE` - Debit from account
    - `WITHDRAWAL` - Debit from account
  - Transaction history
  - PostgreSQL integration
  - gRPC service interface
  - Auto-generated protobuf code

## Internal Packages

### Common Package (internal/common)
- **database.go**: Database connection management and pooling
- **orm.go**: Database models and utilities
- **proto_db.go**: Conversion between protobuf and database models

### Account Package (internal/account)
- **account.go**: Account service business logic
- Implements gRPC AccountService interface
- Handles account operations and validation

### Transaction Package (internal/transaction)
- **transaction.go**: Transaction service business logic
- Implements gRPC TransactionService interface
- Handles transaction processing and validation

### Health Package (internal/health)
- **health.go**: Health check utilities
- Database connectivity checks
- Service health monitoring

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
- Docker (optional)

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
   psql -d pismo -f scripts/database/init.sql
   ```

### Running the Services

#### Option 1: Run Individual Services
1. **Start Account Manager:**
   ```bash
   cd cmd/account-mgr
   go generate  # Generate protobuf code
   go build -o account-mgr .
   ./account-mgr
   ```

2. **Start Transaction Manager:**
   ```bash
   cd cmd/transaction-mgr
   go generate  # Generate protobuf code
   go build -o transaction-mgr .
   ./transaction-mgr
   ```

3. **Start Gateway:**
   ```bash
   cd cmd/gateway
   go build -o gateway .
   PORT=8083 ./gateway
   ```

#### Option 2: Use Test Script
```bash
./test-refactored-system.sh
```

### Docker Support

Each service has its own Dockerfile for containerization:

```bash
# Build and run Account Manager
cd cmd/account-mgr
docker build -t account-mgr .
docker run -p 8081:8081 account-mgr

# Build and run Transaction Manager
cd cmd/transaction-mgr
docker build -t transaction-mgr .
docker run -p 8082:8082 transaction-mgr

# Build and run Gateway
cd cmd/gateway
docker build -t gateway .
docker run -p 8083:8083 gateway
```

## Testing

### Run Comprehensive Tests
```bash
./test-refactored-system.sh
```

### Run Individual Service Tests
```bash
./test-account-service.sh
./test-transaction-service.sh
```

## Key Features

### Architecture Benefits
- **Modular Design**: Separated concerns with internal packages
- **Docker Ready**: Each service can be containerized
- **Auto-generation**: Protobuf code generated automatically
- **Unified Database**: Single PostgreSQL instance
- **ACID Transactions**: Strong consistency across operations
- **gRPC Communication**: High-performance inter-service communication
- **REST API**: Easy integration for external clients

### Code Organization
- **cmd/**: Main applications with go:generate directives
- **internal/**: Private packages for shared functionality
- **proto/**: Protocol buffer definitions
- **Modular go.mod**: Each package has its own dependencies
- **Clean imports**: Proper module structure

### Development Features
- **Hot Reload**: Easy development with go run
- **Dependency Management**: Go modules for each package
- **Code Generation**: Automatic protobuf code generation
- **Health Checks**: Built-in health monitoring
- **Error Handling**: Comprehensive error responses
