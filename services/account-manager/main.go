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
)

type Account struct {
	ID             string
	DocumentNumber string
	AccountType    string
	Balance        float64
	CreatedAt      int64
	UpdatedAt      int64
}

type AccountService struct {
	db *sql.DB
}

func NewAccountService(db *sql.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*CreateAccountResponse, error) {
	if req.DocumentNumber == "" || req.AccountType == "" {
		return &CreateAccountResponse{Error: "missing required fields"}, nil
	}

	id := uuid.New().String()
	now := time.Now().Unix()

	_, err := s.db.Exec(`
		INSERT INTO accounts (id, document_number, account_type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, req.DocumentNumber, req.AccountType, req.InitialBalance, now, now)
	if err != nil {
		log.Printf("db insert failed: %v", err)
		return &CreateAccountResponse{Error: "could not create account"}, nil
	}

	return &CreateAccountResponse{
		Account: &Account{
			ID:             id,
			DocumentNumber: req.DocumentNumber,
			AccountType:    req.AccountType,
			Balance:        req.InitialBalance,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}, nil
}

func (s *AccountService) GetAccount(ctx context.Context, req *GetAccountRequest) (*GetAccountResponse, error) {
	if req.Id == "" {
		return &GetAccountResponse{Error: "id required"}, nil
	}

	var a Account
	err := s.db.QueryRow(`
		SELECT id, document_number, account_type, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`, req.Id).Scan(&a.ID, &a.DocumentNumber, &a.AccountType, &a.Balance, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return &GetAccountResponse{Error: "not found"}, nil
		}
		log.Printf("lookup failed: %v", err)
		return &GetAccountResponse{Error: "db error"}, nil
	}

	return &GetAccountResponse{Account: &a}, nil
}

func (s *AccountService) UpdateAccount(ctx context.Context, req *UpdateAccountRequest) (*UpdateAccountResponse, error) {
	if req.Id == "" {
		return &UpdateAccountResponse{Error: "id required"}, nil
	}

	_, err := s.db.Exec(`
		UPDATE accounts
		SET document_number = COALESCE(NULLIF($2, ''), document_number),
		    account_type    = COALESCE(NULLIF($3, ''), account_type),
		    updated_at      = $4
		WHERE id = $1
	`, req.Id, req.DocumentNumber, req.AccountType, time.Now().Unix())

	if err != nil {
		log.Printf("update failed: %v", err)
		return &UpdateAccountResponse{Error: "could not update"}, nil
	}

	acc, _ := s.GetAccount(ctx, &GetAccountRequest{Id: req.Id})
	return &UpdateAccountResponse{Account: acc.Account}, nil
}

func (s *AccountService) DeleteAccount(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	res, err := s.db.Exec(`DELETE FROM accounts WHERE id = $1`, req.Id)
	if err != nil {
		log.Printf("delete failed: %v", err)
		return &DeleteAccountResponse{Error: "delete error"}, nil
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return &DeleteAccountResponse{Error: "not found"}, nil
	}
	return &DeleteAccountResponse{Success: true}, nil
}

func (s *AccountService) GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error) {
	var bal float64
	err := s.db.QueryRow(`SELECT balance FROM accounts WHERE id = $1`, req.AccountId).Scan(&bal)
	if err != nil {
		if err == sql.ErrNoRows {
			return &GetBalanceResponse{Error: "account not found"}, nil
		}
		return &GetBalanceResponse{Error: "db error"}, nil
	}
	return &GetBalanceResponse{Balance: bal}, nil
}

type CreateAccountRequest struct {
	DocumentNumber string
	AccountType    string
	InitialBalance float64
}
type CreateAccountResponse struct {
	Account *Account
	Error   string
}
type GetAccountRequest struct{ Id string }
type GetAccountResponse struct {
	Account *Account
	Error   string
}
type UpdateAccountRequest struct {
	Id             string
	DocumentNumber string
	AccountType    string
}
type UpdateAccountResponse struct {
	Account *Account
	Error   string
}
type DeleteAccountRequest struct{ Id string }
type DeleteAccountResponse struct {
	Success bool
	Error   string
}
type GetBalanceRequest struct{ AccountId string }
type GetBalanceResponse struct {
	Balance float64
	Error   string
}

func initDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(36) PRIMARY KEY,
			document_number VARCHAR(20) NOT NULL,
			account_type VARCHAR(20) NOT NULL,
			balance DECIMAL(15,2) NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("schema init failed: %w", err)
	}
	return nil
}

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://pismo:pismo123@localhost:5432/pismo?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	if err := initDatabase(db); err != nil {
		log.Fatal(err)
	}

	svc := NewAccountService(db)
	_ = svc // TODO: wire into grpc

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("service on %s", port)
	grpcServer := grpc.NewServer()
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
