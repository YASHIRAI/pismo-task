# Build stage
FROM golang:1.21-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy the entire project
COPY . .

# Set working directory to transaction manager
WORKDIR /app/services/transaction-manager

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transaction-manager .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/transaction-manager/transaction-manager .

# Expose port
EXPOSE 8082

# Run the binary
CMD ["./transaction-manager"]