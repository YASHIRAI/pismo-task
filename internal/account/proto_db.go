package account

import (
	"github.com/YASHIRAI/pismo-task/internal/common"
	pbAccount "github.com/YASHIRAI/pismo-task/proto/account"
)

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
