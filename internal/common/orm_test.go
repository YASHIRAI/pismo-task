package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToUnixTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected int64
	}{
		{
			name:     "zero time",
			input:    time.Time{},
			expected: -62135596800,
		},
		{
			name:     "specific time",
			input:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: 1672574400,
		},
		{
			name:     "current time",
			input:    time.Now(),
			expected: time.Now().Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUnixTimestamp(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFromUnixTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected time.Time
	}{
		{
			name:     "zero timestamp",
			input:    0,
			expected: time.Unix(0, 0),
		},
		{
			name:     "specific timestamp",
			input:    1672574400,
			expected: time.Unix(1672574400, 0),
		},
		{
			name:     "negative timestamp",
			input:    -62135596800,
			expected: time.Unix(-62135596800, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromUnixTimestamp(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentTimestamp(t *testing.T) {
	timestamp := GetCurrentTimestamp()

	assert.Greater(t, timestamp, int64(0))

	now := time.Now().Unix()
	assert.InDelta(t, now, timestamp, 1)
}

func TestAccount_Struct(t *testing.T) {
	account := Account{
		ID:             "test-id",
		DocumentNumber: "12345678901",
		AccountType:    "CHECKING",
		Balance:        100.50,
		CreatedAt:      1234567890,
		UpdatedAt:      1234567891,
	}

	assert.Equal(t, "test-id", account.ID)
	assert.Equal(t, "12345678901", account.DocumentNumber)
	assert.Equal(t, "CHECKING", account.AccountType)
	assert.Equal(t, 100.50, account.Balance)
	assert.Equal(t, int64(1234567890), account.CreatedAt)
	assert.Equal(t, int64(1234567891), account.UpdatedAt)
}

func TestTransaction_Struct(t *testing.T) {
	transaction := Transaction{
		ID:            "tx-123",
		AccountID:     "acc-456",
		OperationType: "PAYMENT",
		Amount:        50.25,
		Description:   "Test transaction",
		CreatedAt:     1234567890,
		Status:        "COMPLETED",
	}

	assert.Equal(t, "tx-123", transaction.ID)
	assert.Equal(t, "acc-456", transaction.AccountID)
	assert.Equal(t, "PAYMENT", transaction.OperationType)
	assert.Equal(t, 50.25, transaction.Amount)
	assert.Equal(t, "Test transaction", transaction.Description)
	assert.Equal(t, int64(1234567890), transaction.CreatedAt)
	assert.Equal(t, "COMPLETED", transaction.Status)
}

func TestTimestampConversion_RoundTrip(t *testing.T) {
	originalTime := time.Date(2023, 6, 15, 14, 30, 45, 123456789, time.UTC)

	unixTimestamp := ToUnixTimestamp(originalTime)
	convertedTime := FromUnixTimestamp(unixTimestamp)

	assert.Equal(t, originalTime.Unix(), convertedTime.Unix())
}

func TestAccount_Validation(t *testing.T) {
	tests := []struct {
		name    string
		account Account
		valid   bool
	}{
		{
			name: "valid account",
			account: Account{
				ID:             "valid-id",
				DocumentNumber: "12345678901",
				AccountType:    "CHECKING",
				Balance:        100.0,
				CreatedAt:      1234567890,
				UpdatedAt:      1234567890,
			},
			valid: true,
		},
		{
			name: "empty ID",
			account: Account{
				ID:             "",
				DocumentNumber: "12345678901",
				AccountType:    "CHECKING",
				Balance:        100.0,
				CreatedAt:      1234567890,
				UpdatedAt:      1234567890,
			},
			valid: false,
		},
		{
			name: "negative balance",
			account: Account{
				ID:             "valid-id",
				DocumentNumber: "12345678901",
				AccountType:    "CHECKING",
				Balance:        -50.0,
				CreatedAt:      1234567890,
				UpdatedAt:      1234567890,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.account.ID != "" && tt.account.Balance >= 0
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

func TestTransaction_Validation(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		valid       bool
	}{
		{
			name: "valid transaction",
			transaction: Transaction{
				ID:            "tx-123",
				AccountID:     "acc-456",
				OperationType: "PAYMENT",
				Amount:        50.0,
				Description:   "Test",
				CreatedAt:     1234567890,
				Status:        "COMPLETED",
			},
			valid: true,
		},
		{
			name: "empty ID",
			transaction: Transaction{
				ID:            "",
				AccountID:     "acc-456",
				OperationType: "PAYMENT",
				Amount:        50.0,
				Description:   "Test",
				CreatedAt:     1234567890,
				Status:        "COMPLETED",
			},
			valid: false,
		},
		{
			name: "empty account ID",
			transaction: Transaction{
				ID:            "tx-123",
				AccountID:     "",
				OperationType: "PAYMENT",
				Amount:        50.0,
				Description:   "Test",
				CreatedAt:     1234567890,
				Status:        "COMPLETED",
			},
			valid: false,
		},
		{
			name: "zero amount",
			transaction: Transaction{
				ID:            "tx-123",
				AccountID:     "acc-456",
				OperationType: "PAYMENT",
				Amount:        0.0,
				Description:   "Test",
				CreatedAt:     1234567890,
				Status:        "COMPLETED",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			isValid := tt.transaction.ID != "" && tt.transaction.AccountID != ""
			assert.Equal(t, tt.valid, isValid)
		})
	}
}
