# build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# deps first for caching
COPY go.mod go.sum ./
RUN go mod download

# app source
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o account-manager

# runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/account-manager .

EXPOSE 8081

CMD ["./account-manager"]
