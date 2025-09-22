package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/YASHIRAI/pismo-task/api/proto/transaction"
)

type Transaction struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"account_id"`
	OperationType string    `json:"operation_type"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"status"`
}

type Account struct {
	ID             string  `json:"id"`
	DocumentNumber string  `json:"document_number"`
	AccountType    string  `json:"account_type"`
	Balance        float64 `json:"balance"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

type TransactionService struct {
	pb.UnimplementedTransactionServiceServer
	db *sql.DB
}

func NewTransactionService(db *sql.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	if req.AccountId == "" || req.OperationType == "" {
		return &pb.CreateTransactionResponse{Error: "missing required fields"}, nil
	}

	// Validate operation type
	validOperations := map[string]bool{
		"CASH_PURCHASE":        true,
		"INSTALLMENT_PURCHASE": true,
		"WITHDRAWAL":           true,
		"PAYMENT":              true,
	}
	if !validOperations[req.OperationType] {
		return &pb.CreateTransactionResponse{Error: "invalid operation type"}, nil
	}

	// Check if account exists
	var account Account
	err := s.db.QueryRow(`
		SELECT id, document_number, account_type, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`, req.AccountId).Scan(&account.ID, &account.DocumentNumber, &account.AccountType, &account.Balance, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.CreateTransactionResponse{Error: "account not found"}, nil
		}
		log.Printf("account check failed: %v", err)
		return &pb.CreateTransactionResponse{Error: "database error"}, nil
	}

	id := uuid.New().String()
	now := time.Now()
	status := "PENDING"

	if req.OperationType == "PAYMENT" {
		if req.Amount <= 0 {
			return &pb.CreateTransactionResponse{Error: "payment amount must be positive"}, nil
		}

		_, err = s.db.Exec(`
			UPDATE accounts 
			SET balance = balance + $1, updated_at = $2 
			WHERE id = $3
		`, req.Amount, now.Unix(), req.AccountId)
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process payment"}, nil
		}
		status = "COMPLETED"
	} else {
	
		amount := req.Amount
		if amount >= 0 {
			amount = -amount
		}

		if account.Balance+amount < 0 {
			return &pb.CreateTransactionResponse{Error: "insufficient balance"}, nil
		}

		_, err = s.db.Exec(`
			UPDATE accounts 
			SET balance = balance + $1, updated_at = $2 
			WHERE id = $3
		`, amount, now.Unix(), req.AccountId)
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process transaction"}, nil
		}
		status = "COMPLETED"
		req.Amount = amount 

	_, err = s.db.Exec(`
		INSERT INTO transactions (id, account_id, operation_type, amount, description, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, req.AccountId, req.OperationType, req.Amount, req.Description, now.Unix(), status)
	if err != nil {
		log.Printf("transaction insert failed: %v", err)
		return &pb.CreateTransactionResponse{Error: "could not create transaction"}, nil
	}

	return &pb.CreateTransactionResponse{
		Transaction: &pb.Transaction{
			Id:            id,
			AccountId:     req.AccountId,
			OperationType: req.OperationType,
			Amount:        req.Amount,
			Description:   req.Description,
			CreatedAt:     now.Unix(),
			Status:        status,
		},
	}, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	if req.Id == "" {
		return &pb.GetTransactionResponse{Error: "id required"}, nil
	}

	var transaction Transaction
	err := s.db.QueryRow(`
		SELECT id, account_id, operation_type, amount, description, created_at, status
		FROM transactions WHERE id = $1
	`, req.Id).Scan(&transaction.ID, &transaction.AccountID, &transaction.OperationType, &transaction.Amount, &transaction.Description, &transaction.CreatedAt, &transaction.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetTransactionResponse{Error: "not found"}, nil
		}
		log.Printf("transaction lookup failed: %v", err)
		return &pb.GetTransactionResponse{Error: "database error"}, nil
	}

	return &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Id:            transaction.ID,
			AccountId:     transaction.AccountID,
			OperationType: transaction.OperationType,
			Amount:        transaction.Amount,
			Description:   transaction.Description,
			CreatedAt:     transaction.CreatedAt.Unix(),
			Status:        transaction.Status,
		},
	}, nil
}

func (s *TransactionService) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	if req.AccountId == "" {
		return &pb.GetTransactionHistoryResponse{Error: "account_id required"}, nil
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var total int32
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM transactions WHERE account_id = $1
	`, req.AccountId).Scan(&total)
	if err != nil {
		log.Printf("count query failed: %v", err)
		return &pb.GetTransactionHistoryResponse{Error: "database error"}, nil
	}

	rows, err := s.db.Query(`
		SELECT id, account_id, operation_type, amount, description, created_at, status
		FROM transactions 
		WHERE account_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`, req.AccountId, limit, offset)
	if err != nil {
		log.Printf("transactions query failed: %v", err)
		return &pb.GetTransactionHistoryResponse{Error: "database error"}, nil
	}
	defer rows.Close()

	var transactions []*pb.Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.AccountID, &t.OperationType, &t.Amount, &t.Description, &t.CreatedAt, &t.Status); err != nil {
			log.Printf("row scan failed: %v", err)
			continue
		}
		transactions = append(transactions, &pb.Transaction{
			Id:            t.ID,
			AccountId:     t.AccountID,
			OperationType: t.OperationType,
			Amount:        t.Amount,
			Description:   t.Description,
			CreatedAt:     t.CreatedAt.Unix(),
			Status:        t.Status,
		})
	}

	return &pb.GetTransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

func (s *TransactionService) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	createReq := &pb.CreateTransactionRequest{
		AccountId:     req.AccountId,
		OperationType: "PAYMENT",
		Amount:        req.Amount,
		Description:   req.Description,
	}

	resp, err := s.CreateTransaction(ctx, createReq)
	if err != nil {
		return &pb.ProcessPaymentResponse{Error: err.Error()}, nil
	}

	return &pb.ProcessPaymentResponse{
		Transaction: resp.Transaction,
		Error:       resp.Error,
	}, nil
}

func initDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(36) PRIMARY KEY,
			account_id VARCHAR(36) NOT NULL,
			operation_type VARCHAR(50) NOT NULL,
			amount DECIMAL(15,2) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("transactions table creation failed: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_transactions_account_created ON transactions(account_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_transactions_operation_type ON transactions(operation_type);
		CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
	`)
	if err != nil {
		log.Printf("index creation failed: %v", err)
	}

	return nil
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://pismo:pismo123@localhost:5432/pismo?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := initDatabase(db); err != nil {
		log.Fatal(err)
	}

	svc := NewTransactionService(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Transaction service listening on port %s", port)
	grpcServer := grpc.NewServer()
	pb.RegisterTransactionServiceServer(grpcServer, svc)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
