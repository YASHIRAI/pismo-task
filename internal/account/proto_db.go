package account

import (
	"github.com/YASHIRAI/pismo-task/internal/common"
	pbAccount "github.com/YASHIRAI/pismo-task/proto/account"
)

// ConvertAccountToProto converts a database Account struct to a protobuf Account message.
// This function maps all fields from the common.Account to the corresponding protobuf fields.
func ConvertAccountToProto(dbAccount *common.Account) *pbAccount.Account {
	return &pbAccount.Account{
		Id:             dbAccount.ID,
		DocumentNumber: dbAccount.DocumentNumber,
		AccountType:    dbAccount.AccountType,
		Balance:        dbAccount.Balance,
		CreatedAt:      dbAccount.CreatedAt,
		UpdatedAt:      dbAccount.UpdatedAt,
	}
}

// ConvertAccountFromProto converts a protobuf Account message to a database Account struct.
// This function maps all fields from the protobuf Account to the corresponding common.Account fields.
func ConvertAccountFromProto(pbAccount *pbAccount.Account) *common.Account {
	return &common.Account{
		ID:             pbAccount.Id,
		DocumentNumber: pbAccount.DocumentNumber,
		AccountType:    pbAccount.AccountType,
		Balance:        pbAccount.Balance,
		CreatedAt:      pbAccount.CreatedAt,
		UpdatedAt:      pbAccount.UpdatedAt,
	}
}

// ConvertCreateAccountRequestToAccount converts a CreateAccountRequest to a database Account struct.
// It sets the current timestamp for both created_at and updated_at fields.
func ConvertCreateAccountRequestToAccount(req *pbAccount.CreateAccountRequest) *common.Account {
	now := common.GetCurrentTimestamp()
	return &common.Account{
		DocumentNumber: req.DocumentNumber,
		AccountType:    req.AccountType,
		Balance:        req.InitialBalance,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
