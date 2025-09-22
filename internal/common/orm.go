package common

import (
	"time"
)

type Account struct {
	ID             string  `db:"id"`
	DocumentNumber string  `db:"document_number"`
	AccountType    string  `db:"account_type"`
	Balance        float64 `db:"balance"`
	CreatedAt      int64   `db:"created_at"`
	UpdatedAt      int64   `db:"updated_at"`
}

type Transaction struct {
	ID            string  `db:"id"`
	AccountID     string  `db:"account_id"`
	OperationType string  `db:"operation_type"`
	Amount        float64 `db:"amount"`
	Description   string  `db:"description"`
	CreatedAt     int64   `db:"created_at"`
	Status        string  `db:"status"`
}

func ToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

func FromUnixTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
