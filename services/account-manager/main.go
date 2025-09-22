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

	pb "github.com/YASHIRAI/pismo-task/api/proto/account"
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
	pb.UnimplementedAccountServiceServer
	db *sql.DB
}

func NewAccountService(db *sql.DB) *AccountService {
	return &AccountService{db: db}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	if req.DocumentNumber == "" || req.AccountType == "" {
		return &pb.CreateAccountResponse{Error: "missing required fields"}, nil
	}

	id := uuid.New().String()
	now := time.Now().Unix()

	_, err := s.db.Exec(`
		INSERT INTO accounts (id, document_number, account_type, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, req.DocumentNumber, req.AccountType, req.InitialBalance, now, now)
	if err != nil {
		log.Printf("db insert failed: %v", err)
		return &pb.CreateAccountResponse{Error: "could not create account"}, nil
	}

	return &pb.CreateAccountResponse{
		Account: &pb.Account{
			Id:             id,
			DocumentNumber: req.DocumentNumber,
			AccountType:    req.AccountType,
			Balance:        req.InitialBalance,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}, nil
}

func (s *AccountService) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	if req.Id == "" {
		return &pb.GetAccountResponse{Error: "id required"}, nil
	}

	var a Account
	err := s.db.QueryRow(`
		SELECT id, document_number, account_type, balance, created_at, updated_at
		FROM accounts WHERE id = $1
	`, req.Id).Scan(&a.ID, &a.DocumentNumber, &a.AccountType, &a.Balance, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetAccountResponse{Error: "not found"}, nil
		}
		log.Printf("lookup failed: %v", err)
		return &pb.GetAccountResponse{Error: "db error"}, nil
	}

	return &pb.GetAccountResponse{Account: &pb.Account{
		Id:             a.ID,
		DocumentNumber: a.DocumentNumber,
		AccountType:    a.AccountType,
		Balance:        a.Balance,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}}, nil
}

func (s *AccountService) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	if req.Id == "" {
		return &pb.UpdateAccountResponse{Error: "id required"}, nil
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
		return &pb.UpdateAccountResponse{Error: "could not update"}, nil
	}

	acc, _ := s.GetAccount(ctx, &pb.GetAccountRequest{Id: req.Id})
	return &pb.UpdateAccountResponse{Account: acc.Account}, nil
}

func (s *AccountService) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	res, err := s.db.Exec(`DELETE FROM accounts WHERE id = $1`, req.Id)
	if err != nil {
		log.Printf("delete failed: %v", err)
		return &pb.DeleteAccountResponse{Error: "delete error"}, nil
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return &pb.DeleteAccountResponse{Error: "not found"}, nil
	}
	return &pb.DeleteAccountResponse{Success: true}, nil
}

func (s *AccountService) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	var bal float64
	err := s.db.QueryRow(`SELECT balance FROM accounts WHERE id = $1`, req.AccountId).Scan(&bal)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetBalanceResponse{Error: "account not found"}, nil
		}
		return &pb.GetBalanceResponse{Error: "db error"}, nil
	}
	return &pb.GetBalanceResponse{Balance: bal}, nil
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Account service listening on port %s", port)
	grpcServer := grpc.NewServer()
	pb.RegisterAccountServiceServer(grpcServer, svc)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
