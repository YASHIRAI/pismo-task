FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY cmd/account-mgr/go.mod cmd/account-mgr/go.sum ./
COPY internal/ ./internal/
COPY proto/ ./proto/

RUN go mod download

COPY cmd/account-mgr/ ./

RUN go generate

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o account-mgr .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/account-mgr .

EXPOSE 8081

CMD ["./account-mgr"]
