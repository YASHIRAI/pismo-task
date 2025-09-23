package common

import (
	"time"
)

// Account represents a bank account in the database.
// It contains all account-related information including balance and metadata.
type Account struct {
	ID             string  `db:"id"`
	DocumentNumber string  `db:"document_number"`
	AccountType    string  `db:"account_type"`
	Balance        float64 `db:"balance"`
	CreatedAt      int64   `db:"created_at"`
	UpdatedAt      int64   `db:"updated_at"`
}

// Transaction represents a financial transaction in the database.
// It contains transaction details including operation type, amount, and status.
type Transaction struct {
	ID            string  `db:"id"`
	AccountID     string  `db:"account_id"`
	OperationType string  `db:"operation_type"`
	Amount        float64 `db:"amount"`
	Description   string  `db:"description"`
	CreatedAt     int64   `db:"created_at"`
	Status        string  `db:"status"`
}

// ToUnixTimestamp converts a time.Time to Unix timestamp (seconds since epoch).
// This is used for storing timestamps in the database as integers.
func ToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

// FromUnixTimestamp converts a Unix timestamp to time.Time.
// This is used for converting database timestamps back to Go time objects.
func FromUnixTimestamp(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// GetCurrentTimestamp returns the current Unix timestamp.
// This provides a consistent way to get the current time as an integer for database operations.
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
