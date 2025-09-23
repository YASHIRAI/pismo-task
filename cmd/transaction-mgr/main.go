//go:generate protoc --go_out=../../proto/transaction --go-grpc_out=../../proto/transaction --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=../../proto/transaction ../../proto/transaction/transaction.proto

package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/YASHIRAI/pismo-task/internal/common"
	"github.com/YASHIRAI/pismo-task/internal/transaction"
	pb "github.com/YASHIRAI/pismo-task/proto/transaction"
)

// main starts the Transaction Manager gRPC service.
// It initializes the database connection, sets up the schema, and starts the gRPC server on port 8082.
// The service handles transaction-related operations including creation, retrieval, and payment processing.
func main() {
	// Initialize logging
	logLevel := common.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logger, err := common.NewLogger("transaction-mgr", logLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting Transaction Manager service")

	dbManager, err := common.NewDatabaseManager()
	if err != nil {
		logger.Fatal("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	logger.Info("Database connection established")

	if err := dbManager.InitSchema(); err != nil {
		logger.Fatal("Failed to initialize database schema: %v", err)
	}

	logger.Info("Database schema initialized")

	transactionService := transaction.NewService(dbManager.GetDB(), logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTransactionServiceServer(grpcServer, transactionService)

	logger.Info("Transaction service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve: %v", err)
	}
}
