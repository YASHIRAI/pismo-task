//go:generate protoc --go_out=../../proto/transaction --go-grpc_out=../../proto/transaction --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=../../proto/transaction ../../proto/transaction/transaction.proto

package main

import (
	"log"
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
	dbManager, err := common.NewDatabaseManager()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	if err := dbManager.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	transactionService := transaction.NewService(dbManager.GetDB())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTransactionServiceServer(grpcServer, transactionService)

	log.Printf("Transaction service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
