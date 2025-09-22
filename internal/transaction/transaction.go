package transaction

import (
	"context"
	"database/sql"
	"log"

	"github.com/YASHIRAI/pismo-task/internal/common"
	pb "github.com/YASHIRAI/pismo-task/proto/transaction"
	"github.com/google/uuid"
)

// Service handles transaction operations
type Service struct {
	pb.UnimplementedTransactionServiceServer
	db *sql.DB
}

// NewService creates a new transaction service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CreateTransaction creates a new transaction
func (s *Service) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
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
	var account common.Account
	err := s.db.QueryRowContext(ctx, `
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

	// Create transaction
	dbTransaction := ConvertCreateTransactionRequestToTransaction(req)
	dbTransaction.ID = uuid.New().String()
	status := "PENDING"

	// Process transaction based on operation type
	if req.OperationType == "PAYMENT" {
		// Payment should be positive amount
		if req.Amount <= 0 {
			return &pb.CreateTransactionResponse{Error: "payment amount must be positive"}, nil
		}

		// Update account balance
		_, err = s.db.ExecContext(ctx, `
			UPDATE accounts 
			SET balance = balance + $1, updated_at = $2 
			WHERE id = $3
		`, req.Amount, common.GetCurrentTimestamp(), req.AccountId)
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process payment"}, nil
		}
		status = "COMPLETED"
	} else {
		// For debit operations, check if account has sufficient balance
		amount := req.Amount
		if amount >= 0 {
			amount = -amount
		}

		// Check if account has sufficient balance
		if account.Balance+amount < 0 {
			return &pb.CreateTransactionResponse{Error: "insufficient balance"}, nil
		}

		// Update account balance
		_, err = s.db.ExecContext(ctx, `
			UPDATE accounts 
			SET balance = balance + $1, updated_at = $2 
			WHERE id = $3
		`, amount, common.GetCurrentTimestamp(), req.AccountId)
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process transaction"}, nil
		}
		status = "COMPLETED"
		dbTransaction.Amount = amount // Update the amount to reflect the negative value
	}

	// Insert transaction record
	dbTransaction.Status = status
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO transactions (id, account_id, operation_type, amount, description, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, dbTransaction.ID, dbTransaction.AccountID, dbTransaction.OperationType, dbTransaction.Amount, dbTransaction.Description, dbTransaction.CreatedAt, dbTransaction.Status)

	if err != nil {
		log.Printf("transaction insert failed: %v", err)
		return &pb.CreateTransactionResponse{Error: "could not create transaction"}, nil
	}

	// Convert back to protobuf
	pbTransaction := ConvertTransactionToProto(dbTransaction)
	return &pb.CreateTransactionResponse{Transaction: pbTransaction}, nil
}

// GetTransaction retrieves a transaction by ID
func (s *Service) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	if req.Id == "" {
		return &pb.GetTransactionResponse{Error: "id required"}, nil
	}

	var dbTransaction common.Transaction
	err := s.db.QueryRowContext(ctx, `
		SELECT id, account_id, operation_type, amount, description, created_at, status
		FROM transactions WHERE id = $1
	`, req.Id).Scan(&dbTransaction.ID, &dbTransaction.AccountID, &dbTransaction.OperationType, &dbTransaction.Amount, &dbTransaction.Description, &dbTransaction.CreatedAt, &dbTransaction.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetTransactionResponse{Error: "not found"}, nil
		}
		log.Printf("transaction lookup failed: %v", err)
		return &pb.GetTransactionResponse{Error: "database error"}, nil
	}

	pbTransaction := ConvertTransactionToProto(&dbTransaction)
	return &pb.GetTransactionResponse{Transaction: pbTransaction}, nil
}

// GetTransactionHistory retrieves transaction history for an account
func (s *Service) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
	if req.AccountId == "" {
		return &pb.GetTransactionHistoryResponse{Error: "account_id required"}, nil
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50 // default limit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Get total count
	var total int32
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM transactions WHERE account_id = $1
	`, req.AccountId).Scan(&total)
	if err != nil {
		log.Printf("count query failed: %v", err)
		return &pb.GetTransactionHistoryResponse{Error: "database error"}, nil
	}

	// Get transactions with pagination
	rows, err := s.db.QueryContext(ctx, `
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
		var dbTransaction common.Transaction
		if err := rows.Scan(&dbTransaction.ID, &dbTransaction.AccountID, &dbTransaction.OperationType, &dbTransaction.Amount, &dbTransaction.Description, &dbTransaction.CreatedAt, &dbTransaction.Status); err != nil {
			log.Printf("row scan failed: %v", err)
			continue
		}
		transactions = append(transactions, ConvertTransactionToProto(&dbTransaction))
	}

	return &pb.GetTransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

// ProcessPayment processes a payment transaction
func (s *Service) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	// ProcessPayment is essentially the same as CreateTransaction with PAYMENT operation type
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
