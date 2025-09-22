#!/bin/bash

echo "=========================================="
echo "Pismo Refactored System Test"
echo "=========================================="
echo "Testing refactored microservices with new structure..."

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' 

check_service() {
    local port=$1
    local service_name=$2
    
    if [ "$service_name" = "Gateway" ]; then
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ $service_name is running on port $port${NC}"
            return 0
        else
            echo -e "${RED}✗ $service_name is not running on port $port${NC}"
            return 1
        fi
    else
        if lsof -i:$port > /dev/null 2>&1; then
            echo -e "${GREEN}✓ $service_name is running on port $port${NC}"
            return 0
        else
            echo -e "${RED}✗ $service_name is not running on port $port${NC}"
            return 1
        fi
    fi
}

start_service() {
    local service_dir=$1
    local service_name=$2
    local port=$3
    
    echo "Starting $service_name..."
    cd "$service_dir"
    
    if [ -f "main.go" ] && grep -q "go:generate" "main.go"; then
        echo "Generating protobuf code for $service_name..."
        go generate
    fi
    
    echo "Building $service_name..."
    go build -o "$service_name" .
    
    if [ "$service_name" = "gateway" ]; then
        PORT=$port ./$service_name &
    else
        ./$service_name &
    fi
    
    local pid=$!
    echo $pid > "/tmp/${service_name}.pid"
    
    sleep 3
    
    if check_service $port "$service_name"; then
        echo -e "${GREEN}✓ $service_name started successfully${NC}"
    else
        echo -e "${RED}✗ $service_name failed to start${NC}"
        return 1
    fi
    
    cd - > /dev/null
}

stop_service() {
    local service_name=$1
    
    if [ -f "/tmp/${service_name}.pid" ]; then
        local pid=$(cat "/tmp/${service_name}.pid")
        if kill -0 $pid 2>/dev/null; then
            echo "Stopping $service_name (PID: $pid)..."
            kill $pid
            rm "/tmp/${service_name}.pid"
        fi
    fi
}

cleanup() {
    echo "Cleaning up services..."
    stop_service "account-mgr"
    stop_service "transaction-mgr"
    stop_service "gateway"
}

trap cleanup EXIT

if ! pgrep -x "postgres" > /dev/null; then
    echo -e "${RED}PostgreSQL is not running. Please start it first:${NC}"
    echo "brew services start postgresql@14"
    exit 1
fi

echo "Initializing database..."
psql -d pismo -f scripts/database/init.sql

echo ""
echo "Starting services..."

start_service "cmd/account-mgr" "account-mgr" "8081"
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to start Account Manager${NC}"
    exit 1
fi

start_service "cmd/transaction-mgr" "transaction-mgr" "8082"
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to start Transaction Manager${NC}"
    exit 1
fi

start_service "cmd/gateway" "gateway" "8083"
if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to start Gateway${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo "Testing API Endpoints"
echo "=========================================="

echo "1. Testing Gateway Health Check..."
if check_service "8083" "Gateway"; then
    echo -e "${GREEN}✓ Gateway health check passed${NC}"
else
    echo -e "${RED}✗ Gateway health check failed${NC}"
    exit 1
fi

echo ""
echo "2. Testing Account Creation..."
ACCOUNT_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"document_number\": \"$(date +%s)\", \"account_type\": \"CHECKING\", \"initial_balance\": 1000}" \
    http://localhost:8083/accounts)

if echo "$ACCOUNT_RESPONSE" | grep -q "id"; then
    ACCOUNT_ID=$(echo "$ACCOUNT_RESPONSE" | jq -r '.id')
    echo -e "${GREEN}✓ Account created successfully${NC}"
    echo "   Account ID: $ACCOUNT_ID"
else
    echo -e "${RED}✗ Account creation failed${NC}"
    echo "   Response: $ACCOUNT_RESPONSE"
    exit 1
fi

echo ""
echo "3. Testing Get Account..."
GET_ACCOUNT_RESPONSE=$(curl -s "http://localhost:8083/accounts/$ACCOUNT_ID")
if echo "$GET_ACCOUNT_RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓ Account retrieved successfully${NC}"
else
    echo -e "${RED}✗ Account retrieval failed${NC}"
    echo "   Response: $GET_ACCOUNT_RESPONSE"
fi

echo ""
echo "4. Testing Get Balance..."
BALANCE_RESPONSE=$(curl -s "http://localhost:8083/accounts/$ACCOUNT_ID/balance")
if echo "$BALANCE_RESPONSE" | grep -q "balance"; then
    echo -e "${GREEN}✓ Balance retrieved successfully${NC}"
    echo "   Balance: $BALANCE_RESPONSE"
else
    echo -e "${RED}✗ Balance retrieval failed${NC}"
    echo "   Response: $BALANCE_RESPONSE"
fi

echo ""
echo "5. Testing Transaction Creation..."
TRANSACTION_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"account_id\": \"$ACCOUNT_ID\", \"operation_type\": \"PAYMENT\", \"amount\": 100.50, \"description\": \"Test payment\"}" \
    http://localhost:8083/transactions)

if echo "$TRANSACTION_RESPONSE" | grep -q "id"; then
    TRANSACTION_ID=$(echo "$TRANSACTION_RESPONSE" | jq -r '.id')
    echo -e "${GREEN}✓ Transaction created successfully${NC}"
    echo "   Transaction ID: $TRANSACTION_ID"
else
    echo -e "${RED}✗ Transaction creation failed${NC}"
    echo "   Response: $TRANSACTION_RESPONSE"
fi

echo ""
echo "6. Testing Transaction History..."
HISTORY_RESPONSE=$(curl -s "http://localhost:8083/accounts/$ACCOUNT_ID/transactions?limit=10&offset=0")
if echo "$HISTORY_RESPONSE" | grep -q "transactions"; then
    echo -e "${GREEN}✓ Transaction history retrieved successfully${NC}"
else
    echo -e "${RED}✗ Transaction history retrieval failed${NC}"
    echo "   Response: $HISTORY_RESPONSE"
fi

echo ""
echo "7. Testing Process Payment..."
PAYMENT_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"account_id\": \"$ACCOUNT_ID\", \"amount\": 50.0, \"description\": \"Test payment\"}" \
    http://localhost:8083/payments)

if echo "$PAYMENT_RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓ Payment processed successfully${NC}"
else
    echo -e "${RED}✗ Payment processing failed${NC}"
    echo "   Response: $PAYMENT_RESPONSE"
fi

