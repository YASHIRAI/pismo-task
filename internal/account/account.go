package account

import (
	"context"
	"database/sql"
	"time"

	"github.com/YASHIRAI/pismo-task/internal/common"
	pb "github.com/YASHIRAI/pismo-task/proto/account"
	"github.com/google/uuid"
)

// Service implements the AccountService gRPC server.
// It handles account-related operations including creation, retrieval, updates, and balance management.
type Service struct {
	pb.UnimplementedAccountServiceServer
	db     *sql.DB
	logger *common.Logger
}

// NewService creates a new instance of the Account service.
// It takes a database connection and logger, and returns a configured Service instance.
func NewService(db *sql.DB, logger *common.Logger) *Service {
	return &Service{db: db, logger: logger}
}

// CreateAccount creates a new account with the provided document number and account type.
// It validates required fields and generates a unique UUID for the account.
// Returns the created account or an error message if creation fails.
func (s *Service) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	s.logger.Info("Creating account: DocumentNumber=%s, AccountType=%s, InitialBalance=%f",
		req.DocumentNumber, req.AccountType, req.InitialBalance)

	if req.DocumentNumber == "" || req.AccountType == "" {
		s.logger.Error("Account creation failed: missing required fields")
		return &pb.CreateAccountResponse{Error: "missing required fields"}, nil
	}

	dbAccount := ConvertCreateAccountRequestToAccount(req)
	dbAccount.ID = uuid.New().String()

	start := time.Now()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO accounts (id, document_number, account_type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, dbAccount.ID, dbAccount.DocumentNumber, dbAccount.AccountType, dbAccount.Balance, dbAccount.CreatedAt, dbAccount.UpdatedAt)
	duration := time.Since(start)

	s.logger.LogDatabase("INSERT", "accounts", duration, err)

	if err != nil {
		s.logger.Error("Account creation failed: %v", err)
		return &pb.CreateAccountResponse{Error: "could not create account"}, nil
	}

	s.logger.Info("Account created successfully: ID=%s", dbAccount.ID)
	pbAccount := ConvertAccountToProto(dbAccount)
	return &pb.CreateAccountResponse{Account: pbAccount}, nil
}

// GetAccount retrieves an account by its ID.
// Returns the account details or an error if the account is not found.
func (s *Service) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	s.logger.Debug("Getting account: ID=%s", req.Id)

	if req.Id == "" {
		s.logger.Error("Get account failed: ID required")
		return &pb.GetAccountResponse{Error: "id required"}, nil
	}

	var dbAccount common.Account
	start := time.Now()
	err := s.db.QueryRowContext(ctx, `
		SELECT id, document_number, account_type, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`, req.Id).Scan(&dbAccount.ID, &dbAccount.DocumentNumber, &dbAccount.AccountType, &dbAccount.Balance, &dbAccount.CreatedAt, &dbAccount.UpdatedAt)
	duration := time.Since(start)

	s.logger.LogDatabase("SELECT", "accounts", duration, err)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Account not found: ID=%s", req.Id)
			return &pb.GetAccountResponse{Error: "not found"}, nil
		}
		s.logger.Error("Account lookup failed: %v", err)
		return &pb.GetAccountResponse{Error: "database error"}, nil
	}

	s.logger.Debug("Account retrieved successfully: ID=%s", dbAccount.ID)
	pbAccount := ConvertAccountToProto(&dbAccount)
	return &pb.GetAccountResponse{Account: pbAccount}, nil
}

// UpdateAccount updates an existing account's document number and/or account type.
// Only non-empty fields are updated, preserving existing values for empty fields.
// Returns the updated account or an error if the update fails.
func (s *Service) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	s.logger.Info("Updating account: ID=%s", req.Id)

	if req.Id == "" {
		s.logger.Error("Update account failed: ID required")
		return &pb.UpdateAccountResponse{Error: "id required"}, nil
	}

	start := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE accounts
		SET document_number = COALESCE(NULLIF($2, ''), document_number),
		    account_type    = COALESCE(NULLIF($3, ''), account_type),
		    updated_at      = $4
		WHERE id = $1
	`, req.Id, req.DocumentNumber, req.AccountType, common.GetCurrentTimestamp())
	duration := time.Since(start)

	s.logger.LogDatabase("UPDATE", "accounts", duration, err)

	if err != nil {
		s.logger.Error("Account update failed: %v", err)
		return &pb.UpdateAccountResponse{Error: "could not update account"}, nil
	}

	s.logger.Info("Account updated successfully: ID=%s", req.Id)
	resp, err := s.GetAccount(ctx, &pb.GetAccountRequest{Id: req.Id})
	if err != nil {
		s.logger.Error("Could not retrieve updated account: %v", err)
		return &pb.UpdateAccountResponse{Error: "could not retrieve updated account"}, nil
	}

	return &pb.UpdateAccountResponse{Account: resp.Account}, nil
}

// DeleteAccount removes an account from the database by its ID.
// Returns success status or an error if the account is not found or deletion fails.
func (s *Service) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	if req.Id == "" {
		return &pb.DeleteAccountResponse{Error: "id required"}, nil
	}

	start := time.Now()
	result, err := s.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, req.Id)
	duration := time.Since(start)

	s.logger.LogDatabase("DELETE", "accounts", duration, err)

	if err != nil {
		s.logger.Error("Account deletion failed: %v", err)
		return &pb.DeleteAccountResponse{Error: "could not delete account"}, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &pb.DeleteAccountResponse{Error: "could not determine deletion result"}, nil
	}

	if rowsAffected == 0 {
		return &pb.DeleteAccountResponse{Error: "account not found"}, nil
	}

	return &pb.DeleteAccountResponse{Success: true}, nil
}

// GetBalance retrieves the current balance of an account by its ID.
// Returns the balance amount or an error if the account is not found.
func (s *Service) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.AccountId == "" {
		return &pb.GetBalanceResponse{Error: "account_id required"}, nil
	}

	var balance float64
	start := time.Now()
	err := s.db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE id = $1`, req.AccountId).Scan(&balance)
	duration := time.Since(start)

	s.logger.LogDatabase("SELECT", "accounts", duration, err)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.Warn("Account not found for balance lookup: ID=%s", req.AccountId)
			return &pb.GetBalanceResponse{Error: "account not found"}, nil
		}
		s.logger.Error("Balance lookup failed: %v", err)
		return &pb.GetBalanceResponse{Error: "database error"}, nil
	}

	return &pb.GetBalanceResponse{Balance: balance}, nil
}
