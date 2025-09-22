package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbAccount "github.com/YASHIRAI/pismo-task/api/proto/account"
	pbTransaction "github.com/YASHIRAI/pismo-task/api/proto/transaction"
)

type GatewayService struct {
	accountClient     pbAccount.AccountServiceClient
	transactionClient pbTransaction.TransactionServiceClient
}

func NewGatewayService(accountConn, transactionConn *grpc.ClientConn) *GatewayService {
	return &GatewayService{
		accountClient:     pbAccount.NewAccountServiceClient(accountConn),
		transactionClient: pbTransaction.NewTransactionServiceClient(transactionConn),
	}
}

// HTTP Handlers
func (g *GatewayService) CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocumentNumber string  `json:"document_number"`
		AccountType    string  `json:"account_type"`
		InitialBalance float64 `json:"initial_balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &pbAccount.CreateAccountRequest{
		DocumentNumber: req.DocumentNumber,
		AccountType:    req.AccountType,
		InitialBalance: req.InitialBalance,
	}

	resp, err := g.accountClient.CreateAccount(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Account service error: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Account)
}

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

func (g *GatewayService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	accountAddr := os.Getenv("ACCOUNT_SERVICE_ADDR")
	if accountAddr == "" {
		accountAddr = "localhost:8081"
	}

	transactionAddr := os.Getenv("TRANSACTION_SERVICE_ADDR")
	if transactionAddr == "" {
		transactionAddr = "localhost:8082"
	}

	accountConn, err := grpc.Dial(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to account service: %v", err)
	}
	defer accountConn.Close()

	transactionConn, err := grpc.Dial(transactionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to transaction service: %v", err)
	}
	defer transactionConn.Close()

	gateway := NewGatewayService(accountConn, transactionConn)

	r := mux.NewRouter()

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
		port = "8080"
	}

	log.Printf("Gateway service listening on port %s", port)
	log.Printf("Account service: %s", accountAddr)
	log.Printf("Transaction service: %s", transactionAddr)

	if err := http.ListenAndServe(":"+port, corsHandler(r)); err != nil {
		log.Fatal(err)
	}
}
