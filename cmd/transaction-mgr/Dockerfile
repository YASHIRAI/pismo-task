FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY cmd/transaction-mgr/go.mod cmd/transaction-mgr/go.sum ./
COPY internal/ ./internal/
COPY proto/ ./proto/

RUN go mod download

COPY cmd/transaction-mgr/ ./

RUN go generate

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transaction-mgr .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Create logs directory
RUN mkdir -p logs

COPY --from=builder /app/transaction-mgr .

EXPOSE 8082

# Set default log level
ENV LOG_LEVEL=INFO

CMD ["./transaction-mgr"]
