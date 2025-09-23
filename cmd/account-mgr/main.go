//go:generate protoc --go_out=../../proto/account --go-grpc_out=../../proto/account --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=../../proto/account ../../proto/account/account.proto

package main

import (
	"log"
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
	dbManager, err := common.NewDatabaseManager()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	if err := dbManager.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	accountService := account.NewService(dbManager.GetDB())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAccountServiceServer(grpcServer, accountService)

	log.Printf("Account service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
