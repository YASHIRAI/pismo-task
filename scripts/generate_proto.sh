#!/bin/bash

# Generate proto files with gRPC Gateway support
# This script generates Go code from proto files including HTTP gateway annotations

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Go to project root
cd "$PROJECT_ROOT"

# Create proto include directory if it doesn't exist
mkdir -p proto/include/google/api

# Download google/api/annotations.proto if it doesn't exist
if [ ! -f "proto/include/google/api/annotations.proto" ]; then
    echo "Downloading google/api/annotations.proto..."
    curl -o proto/include/google/api/annotations.proto \
        https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto
fi

# Download google/api/http.proto if it doesn't exist
if [ ! -f "proto/include/google/api/http.proto" ]; then
    echo "Downloading google/api/http.proto..."
    curl -o proto/include/google/api/http.proto \
        https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto
fi

# Generate proto files for each service
echo "Generating proto files..."

# Account service
echo "Generating account service..."
protoc \
    --proto_path=proto \
    --proto_path=proto/include \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    proto/account/account.proto

# Transaction service
echo "Generating transaction service..."
protoc \
    --proto_path=proto \
    --proto_path=proto/include \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    proto/transaction/transaction.proto

# Health service
echo "Generating health service..."
protoc \
    --proto_path=proto \
    --proto_path=proto/include \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=paths=source_relative \
    proto/health/health.proto

echo "Proto generation completed successfully!"
echo ""
echo "Generated files:"
echo "- proto/account/account.pb.go (updated)"
echo "- proto/account/account_grpc.pb.go (updated)"
echo "- proto/account/account.pb.gw.go (new gateway file)"
echo "- proto/transaction/transaction.pb.go (updated)"
echo "- proto/transaction/transaction_grpc.pb.go (updated)"
echo "- proto/transaction/transaction.pb.gw.go (new gateway file)"
echo "- proto/health/health.pb.go (updated)"
echo "- proto/health/health_grpc.pb.go (updated)"
echo "- proto/health/health.pb.gw.go (new gateway file)"
