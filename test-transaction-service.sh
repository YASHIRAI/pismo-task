#!/bin/bash

# Test script for Transaction Manager Service with MongoDB
echo "Testing Transaction Manager Service with MongoDB..."

# Check if MongoDB is running
if ! pgrep -x "mongod" > /dev/null; then
    echo "MongoDB is not running. Please start MongoDB first:"
    echo "brew services start mongodb-community"
    echo "Or run: mongod --config /usr/local/etc/mongod.conf"
    exit 1
fi

# Create test account in MongoDB first
echo "Creating test account..."
mongosh --eval "
use accounts;
db.accounts.insertOne({
    _id: 'test-account-123',
    document_number: '12345678901',
    account_type: 'CHECKING',
    balance: 1000.00,
    created_at: new Date(),
    updated_at: new Date()
});
" --quiet

# Start the transaction service in background
echo "Starting transaction service..."
cd services/transaction-manager
MONGODB_URI="mongodb://localhost:27017" ./transaction-manager &
TRANSACTION_PID=$!

# Wait for service to start
sleep 3

# Test the service using grpcurl (if available)
if command -v grpcurl &> /dev/null; then
    echo "Testing CreateTransaction (Payment)..."
    grpcurl -plaintext -d '{
        "account_id": "test-account-123",
        "operation_type": "PAGAMENTO",
        "amount": 100.50,
        "description": "Test payment"
    }' localhost:8082 transaction.TransactionService/CreateTransaction

    echo "Testing CreateTransaction (Purchase)..."
    grpcurl -plaintext -d '{
        "account_id": "test-account-123",
        "operation_type": "COMPRA_A_VISTA",
        "amount": 50.25,
        "description": "Test purchase"
    }' localhost:8082 transaction.TransactionService/CreateTransaction

    echo "Testing GetTransactionHistory..."
    grpcurl -plaintext -d '{
        "account_id": "test-account-123",
        "limit": 10,
        "offset": 0
    }' localhost:8082 transaction.TransactionService/GetTransactionHistory

    echo "Testing ProcessPayment..."
    grpcurl -plaintext -d '{
        "account_id": "test-account-123",
        "amount": 25.00,
        "description": "Another payment"
    }' localhost:8082 transaction.TransactionService/ProcessPayment

    echo "Checking final account balance..."
    mongosh --eval "
    use accounts;
    db.accounts.findOne({_id: 'test-account-123'}, {balance: 1, updated_at: 1});
    " --quiet
else
    echo "grpcurl not found. Please install it to test the service:"
    echo "go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

# Clean up
echo "Stopping transaction service..."
kill $TRANSACTION_PID 2>/dev/null

echo "Transaction service test completed."
