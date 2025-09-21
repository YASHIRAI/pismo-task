package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/YASHIRAI/pismo-task/api/proto/transaction"
)

type Transaction struct {
	ID            string    `bson:"_id,omitempty"`
	AccountID     string    `bson:"account_id"`
	OperationType string    `bson:"operation_type"`
	Amount        float64   `bson:"amount"`
	Description   string    `bson:"description"`
	CreatedAt     time.Time `bson:"created_at"`
	Status        string    `bson:"status"`
}

type Account struct {
	ID             string    `bson:"_id,omitempty"`
	DocumentNumber string    `bson:"document_number"`
	AccountType    string    `bson:"account_type"`
	Balance        float64   `bson:"balance"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

type TransactionService struct {
	pb.UnimplementedTransactionServiceServer
	client         *mongo.Client
	transactionsDB *mongo.Database
	accountsDB     *mongo.Database
}

func NewTransactionService(client *mongo.Client, transactionsDB, accountsDB *mongo.Database) *TransactionService {
	return &TransactionService{
		client:         client,
		transactionsDB: transactionsDB,
		accountsDB:     accountsDB,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	if req.AccountId == "" || req.OperationType == "" {
		return &pb.CreateTransactionResponse{Error: "missing required fields"}, nil
	}

	// Validate operation type
	validOperations := map[string]bool{
		"COMPRA_A_VISTA":   true,
		"COMPRA_PARCELADA": true,
		"SAQUE":            true,
		"PAGAMENTO":        true,
	}
	if !validOperations[req.OperationType] {
		return &pb.CreateTransactionResponse{Error: "invalid operation type"}, nil
	}

	// Check if account exists
	accountsCollection := s.accountsDB.Collection("accounts")
	var account Account
	err := accountsCollection.FindOne(ctx, bson.M{"_id": req.AccountId}).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.CreateTransactionResponse{Error: "account not found"}, nil
		}
		log.Printf("account check failed: %v", err)
		return &pb.CreateTransactionResponse{Error: "database error"}, nil
	}

	id := uuid.New().String()
	now := time.Now()
	status := "PENDING"

	// For payment operations, we need to check balance
	if req.OperationType == "PAGAMENTO" {
		// Payment should be positive amount
		if req.Amount <= 0 {
			return &pb.CreateTransactionResponse{Error: "payment amount must be positive"}, nil
		}

		// Update account balance
		_, err = accountsCollection.UpdateOne(ctx,
			bson.M{"_id": req.AccountId},
			bson.M{
				"$inc": bson.M{"balance": req.Amount},
				"$set": bson.M{"updated_at": now},
			})
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process payment"}, nil
		}
		status = "COMPLETED"
	} else {
		// For debit operations, check if account has sufficient balance
		// Debit operations should be negative amount
		amount := req.Amount
		if amount >= 0 {
			amount = -amount
		}

		// Check if account has sufficient balance
		if account.Balance+amount < 0 {
			return &pb.CreateTransactionResponse{Error: "insufficient balance"}, nil
		}

		// Update account balance
		_, err = accountsCollection.UpdateOne(ctx,
			bson.M{"_id": req.AccountId},
			bson.M{
				"$inc": bson.M{"balance": amount},
				"$set": bson.M{"updated_at": now},
			})
		if err != nil {
			log.Printf("balance update failed: %v", err)
			return &pb.CreateTransactionResponse{Error: "could not process transaction"}, nil
		}
		status = "COMPLETED"
		req.Amount = amount // Update the amount to reflect the negative value
	}

	// Insert transaction record
	transactionsCollection := s.transactionsDB.Collection("transactions")
	transaction := Transaction{
		ID:            id,
		AccountID:     req.AccountId,
		OperationType: req.OperationType,
		Amount:        req.Amount,
		Description:   req.Description,
		CreatedAt:     now,
		Status:        status,
	}

	_, err = transactionsCollection.InsertOne(ctx, transaction)
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

	transactionsCollection := s.transactionsDB.Collection("transactions")
	var transaction Transaction
	err := transactionsCollection.FindOne(ctx, bson.M{"_id": req.Id}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
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
		limit = 50 // default limit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	transactionsCollection := s.transactionsDB.Collection("transactions")

	// Get total count
	total, err := transactionsCollection.CountDocuments(ctx, bson.M{"account_id": req.AccountId})
	if err != nil {
		log.Printf("count query failed: %v", err)
		return &pb.GetTransactionHistoryResponse{Error: "database error"}, nil
	}

	// Get transactions with pagination
	opts := options.Find().
		SetSort(bson.D{{"created_at", -1}}). // Sort by created_at descending
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := transactionsCollection.Find(ctx, bson.M{"account_id": req.AccountId}, opts)
	if err != nil {
		log.Printf("transactions query failed: %v", err)
		return &pb.GetTransactionHistoryResponse{Error: "database error"}, nil
	}
	defer cursor.Close(ctx)

	var transactions []*pb.Transaction
	for cursor.Next(ctx) {
		var t Transaction
		if err := cursor.Decode(&t); err != nil {
			log.Printf("row decode failed: %v", err)
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
		Total:        int32(total),
	}, nil
}

func (s *TransactionService) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	// ProcessPayment is essentially the same as CreateTransaction with PAGAMENTO operation type
	createReq := &pb.CreateTransactionRequest{
		AccountId:     req.AccountId,
		OperationType: "PAGAMENTO",
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

func initDatabase(client *mongo.Client, transactionsDB, accountsDB *mongo.Database) error {
	// Create indexes for better performance
	transactionsCollection := transactionsDB.Collection("transactions")
	accountsCollection := accountsDB.Collection("accounts")

	// Create indexes for transactions
	_, err := transactionsCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{"account_id", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
		{
			Keys: bson.D{{"account_id", 1}, {"created_at", -1}},
		},
	})
	if err != nil {
		log.Printf("transactions index creation failed: %v", err)
	}

	// Create indexes for accounts
	_, err = accountsCollection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{{"document_number", 1}},
		},
		{
			Keys: bson.D{{"account_type", 1}},
		},
	})
	if err != nil {
		log.Printf("accounts index creation failed: %v", err)
	}

	return nil
}

func main() {
	// MongoDB connection
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// Test the connection
	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	// Get databases
	transactionsDB := client.Database("transactions")
	accountsDB := client.Database("accounts")

	// Initialize database indexes
	if err := initDatabase(client, transactionsDB, accountsDB); err != nil {
		log.Fatal(err)
	}

	svc := NewTransactionService(client, transactionsDB, accountsDB)

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
