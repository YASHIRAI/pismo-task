# Pismo Microservices Task

This repository contains the initial structure for a microservices architecture implementation for a financial services platform.

## Current Implementation

### gRPC Service Definitions

The repository currently contains Protocol Buffer definitions for the planned microservices:

#### Account Service (`api/proto/account.proto`)
Defines the Account Service interface with the following RPC methods:

- `CreateAccount` - Create new accounts with document number, type, and initial balance
- `GetAccount` - Retrieve account details by ID
- `UpdateAccount` - Update account information
- `DeleteAccount` - Remove accounts
- `GetBalance` - Get current account balance

**Account Message Fields:**
- `id` - Unique account identifier
- `document_number` - Customer document number
- `account_type` - Type of account (e.g., "checking", "savings")
- `balance` - Current account balance
- `created_at` - Account creation timestamp
- `updated_at` - Last update timestamp

#### Transaction Service (`api/proto/transaction.proto`)
Defines the Transaction Service interface with the following RPC methods:

- `CreateTransaction` - Create new transactions
- `GetTransaction` - Retrieve transaction details by ID
- `GetTransactionHistory` - Get paginated transaction history for an account
- `ProcessPayment` - Process payment transactions

**Transaction Message Fields:**
- `id` - Unique transaction identifier
- `account_id` - Associated account ID
- `operation_type` - Type of operation (e.g., "debit", "credit", "transfer")
- `amount` - Transaction amount
- `description` - Transaction description
- `created_at` - Transaction creation timestamp
- `status` - Transaction status (e.g., "pending", "completed", "failed")

## Planned Architecture

The system is designed to consist of the following microservices:

- **Gateway**: API gateway that routes requests to appropriate services
- **Account Manager**: Handles account creation, management, and balance operations
- **Transaction Manager**: Processes financial transactions and maintains transaction history
- **Database**: PostgreSQL database for data persistence

## Development Setup

1. **Prerequisites**
   - Go 1.21+ (for local development)
   - Protocol Buffers compiler (`protoc`)

2. **Generate gRPC Code** (when implementing services)
   ```bash
   # Generate Go code from proto files
   protoc --go_out=. --go-grpc_out=. api/proto/*.proto
   ```

## Project Structure

```
├── api/
│   └── proto/           # Protocol buffer definitions
│       ├── account.proto    # Account service definitions
│       └── transaction.proto # Transaction service definitions
├── internal/            # Shared internal packages (placeholder)
├── services/            # Microservice implementations (placeholder)
├── tests/               # Test files (placeholder)
├── docker-compose.yml   # Service orchestration (placeholder)
└── test-microservices.sh # Test runner script (placeholder)
```
