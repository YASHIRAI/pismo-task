package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/YASHIRAI/pismo-task/internal/common"
	pbAccount "github.com/YASHIRAI/pismo-task/proto/account"
	pbTransaction "github.com/YASHIRAI/pismo-task/proto/transaction"
)

// GatewayService provides HTTP REST API endpoints that route requests to gRPC services.
// It acts as a gateway between external clients and the internal microservices.
type GatewayService struct {
	accountClient     pbAccount.AccountServiceClient
	transactionClient pbTransaction.TransactionServiceClient
	logger            *common.Logger
}

// LoggingMiddleware provides HTTP request logging functionality
func LoggingMiddleware(logger *common.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(wrapped, r)

			// Log the request
			duration := time.Since(start)
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			}

			logger.LogRequest(r.Method, r.URL.Path, clientIP, wrapped.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewGatewayService creates a new gateway service instance.
// It takes gRPC client connections for account and transaction services and returns a configured GatewayService.
func NewGatewayService(accountConn, transactionConn *grpc.ClientConn, logger *common.Logger) *GatewayService {
	return &GatewayService{
		accountClient:     pbAccount.NewAccountServiceClient(accountConn),
		transactionClient: pbTransaction.NewTransactionServiceClient(transactionConn),
		logger:            logger,
	}
}

// CreateAccountHandler handles HTTP POST requests to create new accounts.
// It accepts JSON input, converts it to gRPC format, and returns the created account or error.
func (g *GatewayService) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	g.logger.Info("Creating new account")

	var req struct {
		DocumentNumber string  `json:"document_number"`
		AccountType    string  `json:"account_type"`
		InitialBalance float64 `json:"initial_balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.logger.Error("Failed to decode JSON request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	g.logger.Debug("Account creation request: DocumentNumber=%s, AccountType=%s, InitialBalance=%f",
		req.DocumentNumber, req.AccountType, req.InitialBalance)

	grpcReq := &pbAccount.CreateAccountRequest{
		DocumentNumber: req.DocumentNumber,
		AccountType:    req.AccountType,
		InitialBalance: req.InitialBalance,
	}

	start := time.Now()
	resp, err := g.accountClient.CreateAccount(context.Background(), grpcReq)
	duration := time.Since(start)

	g.logger.LogGRPC("CreateAccount", duration, err)

	if err != nil {
		g.logger.Error("Account service error: %v", err)
		http.Error(w, fmt.Sprintf("Account service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		g.logger.Error("Account creation failed: %s", resp.Error)
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	g.logger.Info("Account created successfully: ID=%s", resp.Account.Id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Account)
}

// GetAccountHandler handles HTTP GET requests to retrieve account details by ID.
// It extracts the account ID from the URL path and returns the account information or error.
func (g *GatewayService) GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	grpcReq := &pbAccount.GetAccountRequest{Id: accountID}
	resp, err := g.accountClient.GetAccount(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Account service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Account)
}

// GetBalanceHandler handles HTTP GET requests to retrieve account balance by ID.
// It extracts the account ID from the URL path and returns the current balance or error.
func (g *GatewayService) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["id"]

	grpcReq := &pbAccount.GetBalanceRequest{AccountId: accountID}
	resp, err := g.accountClient.GetBalance(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Account service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"balance": resp.Balance})
}

// CreateTransactionHandler handles HTTP POST requests to create new transactions.
// It accepts JSON input, converts it to gRPC format, and returns the created transaction or error.
func (g *GatewayService) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID     string  `json:"account_id"`
		OperationType string  `json:"operation_type"`
		Amount        float64 `json:"amount"`
		Description   string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pbTransaction.CreateTransactionRequest{
		AccountId:     req.AccountID,
		OperationType: req.OperationType,
		Amount:        req.Amount,
		Description:   req.Description,
	}

	resp, err := g.transactionClient.CreateTransaction(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Transaction)
}

// GetTransactionHandler handles HTTP GET requests to retrieve transaction details by ID.
// It extracts the transaction ID from the URL path and returns the transaction information or error.
func (g *GatewayService) GetTransactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionID := vars["id"]

	grpcReq := &pbTransaction.GetTransactionRequest{Id: transactionID}
	resp, err := g.transactionClient.GetTransaction(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Transaction)
}

// GetTransactionHistoryHandler handles HTTP GET requests to retrieve transaction history for an account.
// It supports pagination with limit and offset query parameters and returns the transaction list with total count.
func (g *GatewayService) GetTransactionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["account_id"]

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(50)
	offset := int32(0)

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(l)
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = int32(o)
		}
	}

	grpcReq := &pbTransaction.GetTransactionHistoryRequest{
		AccountId: accountID,
		Limit:     limit,
		Offset:    offset,
	}

	resp, err := g.transactionClient.GetTransactionHistory(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"transactions": resp.Transactions,
		"total":        resp.Total,
	})
}

// ProcessPaymentHandler handles HTTP POST requests to process payment transactions.
// It accepts JSON input for payment details and returns the processed transaction or error.
func (g *GatewayService) ProcessPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountID   string  `json:"account_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pbTransaction.ProcessPaymentRequest{
		AccountId:   req.AccountID,
		Amount:      req.Amount,
		Description: req.Description,
	}

	resp, err := g.transactionClient.ProcessPayment(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Transaction)
}

// HealthHandler handles HTTP GET requests for health checks.
// It returns the current service status and timestamp in JSON format.
func (g *GatewayService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// main starts the Gateway HTTP service.
// It establishes connections to account and transaction gRPC services, sets up HTTP routes,
// configures CORS, and starts the HTTP server on port 8080 (or PORT environment variable).
func main() {
	logLevel := common.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logger, err := common.NewLogger("gateway", logLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting Gateway service")

	accountAddr := os.Getenv("ACCOUNT_SERVICE_ADDR")
	if accountAddr == "" {
		accountAddr = "localhost:8081"
	}

	transactionAddr := os.Getenv("TRANSACTION_SERVICE_ADDR")
	if transactionAddr == "" {
		transactionAddr = "localhost:8082"
	}

	logger.Info("Connecting to services: Account=%s, Transaction=%s", accountAddr, transactionAddr)

	accountConn, err := grpc.Dial(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to account service: %v", err)
	}
	defer accountConn.Close()

	transactionConn, err := grpc.Dial(transactionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to transaction service: %v", err)
	}
	defer transactionConn.Close()

	logger.Info("Successfully connected to all services")

	gateway := NewGatewayService(accountConn, transactionConn, logger)

	r := mux.NewRouter()

	// Add logging middleware
	r.Use(LoggingMiddleware(logger))

	r.HandleFunc("/health", gateway.HealthHandler).Methods("GET")

	r.HandleFunc("/accounts", gateway.CreateAccountHandler).Methods("POST")
	r.HandleFunc("/accounts/{id}", gateway.GetAccountHandler).Methods("GET")
	r.HandleFunc("/accounts/{id}/balance", gateway.GetBalanceHandler).Methods("GET")

	r.HandleFunc("/transactions", gateway.CreateTransactionHandler).Methods("POST")
	r.HandleFunc("/transactions/{id}", gateway.GetTransactionHandler).Methods("GET")
	r.HandleFunc("/accounts/{account_id}/transactions", gateway.GetTransactionHistoryHandler).Methods("GET")
	r.HandleFunc("/payments", gateway.ProcessPaymentHandler).Methods("POST")

	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Info("Gateway service listening on port %s", port)
	logger.Info("Account service: %s", accountAddr)
	logger.Info("Transaction service: %s", transactionAddr)

	if err := http.ListenAndServe(":"+port, corsHandler(r)); err != nil {
		logger.Fatal("HTTP server error: %v", err)
	}
}
