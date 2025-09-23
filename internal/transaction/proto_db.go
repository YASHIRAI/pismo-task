package transaction

import (
	"github.com/YASHIRAI/pismo-task/internal/common"
	pbTransaction "github.com/YASHIRAI/pismo-task/proto/transaction"
)

// ConvertTransactionToProto converts a database Transaction struct to a protobuf Transaction message.
// This function maps all fields from the common.Transaction to the corresponding protobuf fields.
func ConvertTransactionToProto(dbTransaction *common.Transaction) *pbTransaction.Transaction {
	return &pbTransaction.Transaction{
		Id:            dbTransaction.ID,
		AccountId:     dbTransaction.AccountID,
		OperationType: dbTransaction.OperationType,
		Amount:        dbTransaction.Amount,
		Description:   dbTransaction.Description,
		CreatedAt:     dbTransaction.CreatedAt,
		Status:        dbTransaction.Status,
	}
}

// ConvertTransactionFromProto converts a protobuf Transaction message to a database Transaction struct.
// This function maps all fields from the protobuf Transaction to the corresponding common.Transaction fields.
func ConvertTransactionFromProto(pbTransaction *pbTransaction.Transaction) *common.Transaction {
	return &common.Transaction{
		ID:            pbTransaction.Id,
		AccountID:     pbTransaction.AccountId,
		OperationType: pbTransaction.OperationType,
		Amount:        pbTransaction.Amount,
		Description:   pbTransaction.Description,
		CreatedAt:     pbTransaction.CreatedAt,
		Status:        pbTransaction.Status,
	}
}

// ConvertCreateTransactionRequestToTransaction converts a CreateTransactionRequest to a database Transaction struct.
// It sets the current timestamp for the created_at field and initializes status as PENDING.
func ConvertCreateTransactionRequestToTransaction(req *pbTransaction.CreateTransactionRequest) *common.Transaction {
	now := common.GetCurrentTimestamp()
	return &common.Transaction{
		AccountID:     req.AccountId,
		OperationType: req.OperationType,
		Amount:        req.Amount,
		Description:   req.Description,
		CreatedAt:     now,
		Status:        "PENDING",
	}
}

// ConvertProcessPaymentRequestToTransaction converts a ProcessPaymentRequest to a database Transaction struct.
// It sets the operation type to PAYMENT, uses the current timestamp, and initializes status as PENDING.
func ConvertProcessPaymentRequestToTransaction(req *pbTransaction.ProcessPaymentRequest) *common.Transaction {
	now := common.GetCurrentTimestamp()
	return &common.Transaction{
		AccountID:     req.AccountId,
		OperationType: "PAYMENT",
		Amount:        req.Amount,
		Description:   req.Description,
		CreatedAt:     now,
		Status:        "PENDING",
	}
}
