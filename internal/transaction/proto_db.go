package transaction

import (
	"github.com/YASHIRAI/pismo-task/internal/common"
	pbTransaction "github.com/YASHIRAI/pismo-task/proto/transaction"
)

// ConvertTransactionToProto converts database Transaction to protobuf Transaction
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

// ConvertTransactionFromProto converts protobuf Transaction to database Transaction
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

// ConvertCreateTransactionRequestToTransaction converts CreateTransactionRequest to database Transaction
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

// ConvertProcessPaymentRequestToTransaction converts ProcessPaymentRequest to database Transaction
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
