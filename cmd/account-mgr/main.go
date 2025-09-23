//go:generate protoc --go_out=../../proto/account --go-grpc_out=../../proto/account --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=../../proto/account ../../proto/account/account.proto

package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/YASHIRAI/pismo-task/internal/account"
	"github.com/YASHIRAI/pismo-task/internal/common"
	pb "github.com/YASHIRAI/pismo-task/proto/account"
)

// main starts the Account Manager gRPC service.
// It initializes the database connection, sets up the schema, and starts the gRPC server on port 8081.
// The service handles account-related operations including CRUD operations and balance management.
func main() {
	// Initialize logging
	logLevel := common.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logger, err := common.NewLogger("account-mgr", logLevel)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting Account Manager service")

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

	accountService := account.NewService(dbManager.GetDB(), logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAccountServiceServer(grpcServer, accountService)

	logger.Info("Account service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve: %v", err)
	}
}
