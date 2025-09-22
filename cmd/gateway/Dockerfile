FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY cmd/gateway/go.mod cmd/gateway/go.sum ./
COPY proto/ ./proto/

RUN go mod download

COPY cmd/gateway/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/gateway .

EXPOSE 8080

CMD ["./gateway"]
