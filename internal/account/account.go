package account

import (
	"context"
	"database/sql"
	"log"

	"github.com/YASHIRAI/pismo-task/internal/common"
	pb "github.com/YASHIRAI/pismo-task/proto/account"
	"github.com/google/uuid"
)

type Service struct {
	pb.UnimplementedAccountServiceServer
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	if req.DocumentNumber == "" || req.AccountType == "" {
		return &pb.CreateAccountResponse{Error: "missing required fields"}, nil
	}

	dbAccount := ConvertCreateAccountRequestToAccount(req)
	dbAccount.ID = uuid.New().String()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO accounts (id, document_number, account_type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, dbAccount.ID, dbAccount.DocumentNumber, dbAccount.AccountType, dbAccount.Balance, dbAccount.CreatedAt, dbAccount.UpdatedAt)

	if err != nil {
		log.Printf("account creation failed: %v", err)
		return &pb.CreateAccountResponse{Error: "could not create account"}, nil
	}

	pbAccount := ConvertAccountToProto(dbAccount)
	return &pb.CreateAccountResponse{Account: pbAccount}, nil
}

func (s *Service) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	if req.Id == "" {
		return &pb.GetAccountResponse{Error: "id required"}, nil
	}

	var dbAccount common.Account
	err := s.db.QueryRowContext(ctx, `
		SELECT id, document_number, account_type, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`, req.Id).Scan(&dbAccount.ID, &dbAccount.DocumentNumber, &dbAccount.AccountType, &dbAccount.Balance, &dbAccount.CreatedAt, &dbAccount.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetAccountResponse{Error: "not found"}, nil
		}
		log.Printf("account lookup failed: %v", err)
		return &pb.GetAccountResponse{Error: "database error"}, nil
	}

	pbAccount := ConvertAccountToProto(&dbAccount)
	return &pb.GetAccountResponse{Account: pbAccount}, nil
}

func (s *Service) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	if req.Id == "" {
		return &pb.UpdateAccountResponse{Error: "id required"}, nil
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE accounts
		SET document_number = COALESCE(NULLIF($2, ''), document_number),
		    account_type    = COALESCE(NULLIF($3, ''), account_type),
		    updated_at      = $4
		WHERE id = $1
	`, req.Id, req.DocumentNumber, req.AccountType, common.GetCurrentTimestamp())

	if err != nil {
		log.Printf("account update failed: %v", err)
		return &pb.UpdateAccountResponse{Error: "could not update account"}, nil
	}

	resp, err := s.GetAccount(ctx, &pb.GetAccountRequest{Id: req.Id})
	if err != nil {
		return &pb.UpdateAccountResponse{Error: "could not retrieve updated account"}, nil
	}

	return &pb.UpdateAccountResponse{Account: resp.Account}, nil
}

func (s *Service) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	if req.Id == "" {
		return &pb.DeleteAccountResponse{Error: "id required"}, nil
	}

	result, err := s.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, req.Id)
	if err != nil {
		log.Printf("account deletion failed: %v", err)
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

func (s *Service) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.AccountId == "" {
		return &pb.GetBalanceResponse{Error: "account_id required"}, nil
	}

	var balance float64
	err := s.db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE id = $1`, req.AccountId).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetBalanceResponse{Error: "account not found"}, nil
		}
		log.Printf("balance lookup failed: %v", err)
		return &pb.GetBalanceResponse{Error: "database error"}, nil
	}

	return &pb.GetBalanceResponse{Balance: balance}, nil
}
