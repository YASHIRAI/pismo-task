package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	pb "github.com/YASHIRAI/pismo-task/proto/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	service := NewService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestService_CreateTransaction(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.CreateTransactionRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.CreateTransactionResponse
	}{
		{
			name: "successful payment transaction",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "PAYMENT",
				Amount:        100.50,
				Description:   "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 200.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)

				// Mock balance update
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs(100.50, sqlmock.AnyArg(), "test-account-id").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock transaction insert
				mock.ExpectExec(`INSERT INTO transactions`).
					WithArgs(sqlmock.AnyArg(), "test-account-id", "PAYMENT", 100.50, "Test payment", sqlmock.AnyArg(), "COMPLETED").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
			expectedResult: &pb.CreateTransactionResponse{
				Transaction: &pb.Transaction{
					AccountId:     "test-account-id",
					OperationType: "PAYMENT",
					Amount:        100.50,
					Description:   "Test payment",
					Status:        "COMPLETED",
				},
			},
		},
		{
			name: "successful cash purchase transaction",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "CASH_PURCHASE",
				Amount:        50.00,
				Description:   "Test purchase",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 200.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)

				// Mock balance update (negative amount)
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs(-50.00, sqlmock.AnyArg(), "test-account-id").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock transaction insert
				mock.ExpectExec(`INSERT INTO transactions`).
					WithArgs(sqlmock.AnyArg(), "test-account-id", "CASH_PURCHASE", -50.00, "Test purchase", sqlmock.AnyArg(), "COMPLETED").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
			expectedResult: &pb.CreateTransactionResponse{
				Transaction: &pb.Transaction{
					AccountId:     "test-account-id",
					OperationType: "CASH_PURCHASE",
					Amount:        -50.00,
					Description:   "Test purchase",
					Status:        "COMPLETED",
				},
			},
		},
		{
			name: "missing required fields",
			request: &pb.CreateTransactionRequest{
				AccountId:     "",
				OperationType: "PAYMENT",
				Amount:        100.50,
				Description:   "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "missing required fields",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "missing required fields",
			},
		},
		{
			name: "invalid operation type",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "INVALID_OPERATION",
				Amount:        100.50,
				Description:   "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "invalid operation type",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "invalid operation type",
			},
		},
		{
			name: "account not found",
			request: &pb.CreateTransactionRequest{
				AccountId:     "non-existent-id",
				OperationType: "PAYMENT",
				Amount:        100.50,
				Description:   "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("non-existent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "account not found",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "account not found",
			},
		},
		{
			name: "insufficient balance for debit operation",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "CASH_PURCHASE",
				Amount:        500.00,
				Description:   "Large purchase",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup with low balance
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 100.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)
			},
			expectedError: "insufficient balance",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "insufficient balance",
			},
		},
		{
			name: "negative payment amount",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "PAYMENT",
				Amount:        -100.50,
				Description:   "Invalid payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 200.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)
			},
			expectedError: "payment amount must be positive",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "payment amount must be positive",
			},
		},
		{
			name: "database error during account lookup",
			request: &pb.CreateTransactionRequest{
				AccountId:     "test-account-id",
				OperationType: "PAYMENT",
				Amount:        100.50,
				Description:   "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "database error",
			expectedResult: &pb.CreateTransactionResponse{
				Error: "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			service := NewService(db)
			response, err := service.CreateTransaction(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedError == "" {
				assert.NotEmpty(t, response.Transaction.Id)
				assert.Equal(t, tt.request.AccountId, response.Transaction.AccountId)
				assert.Equal(t, tt.request.OperationType, response.Transaction.OperationType)
				assert.Equal(t, tt.request.Description, response.Transaction.Description)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetTransaction(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.GetTransactionRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.GetTransactionResponse
	}{
		{
			name: "successful transaction retrieval",
			request: &pb.GetTransactionRequest{
				Id: "test-transaction-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "account_id", "operation_type", "amount", "description", "created_at", "status"}).
					AddRow("test-transaction-id", "test-account-id", "PAYMENT", 100.50, "Test payment", 1234567890, "COMPLETED")
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("test-transaction-id").
					WillReturnRows(rows)
			},
			expectedError: "",
			expectedResult: &pb.GetTransactionResponse{
				Transaction: &pb.Transaction{
					Id:            "test-transaction-id",
					AccountId:     "test-account-id",
					OperationType: "PAYMENT",
					Amount:        100.50,
					Description:   "Test payment",
					CreatedAt:     1234567890,
					Status:        "COMPLETED",
				},
			},
		},
		{
			name: "missing transaction id",
			request: &pb.GetTransactionRequest{
				Id: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "id required",
			expectedResult: &pb.GetTransactionResponse{
				Error: "id required",
			},
		},
		{
			name: "transaction not found",
			request: &pb.GetTransactionRequest{
				Id: "non-existent-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("non-existent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "not found",
			expectedResult: &pb.GetTransactionResponse{
				Error: "not found",
			},
		},
		{
			name: "database error",
			request: &pb.GetTransactionRequest{
				Id: "test-transaction-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("test-transaction-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "database error",
			expectedResult: &pb.GetTransactionResponse{
				Error: "database error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			service := NewService(db)
			response, err := service.GetTransaction(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedError == "" {
				assert.Equal(t, tt.expectedResult.Transaction.Id, response.Transaction.Id)
				assert.Equal(t, tt.expectedResult.Transaction.AccountId, response.Transaction.AccountId)
				assert.Equal(t, tt.expectedResult.Transaction.OperationType, response.Transaction.OperationType)
				assert.Equal(t, tt.expectedResult.Transaction.Amount, response.Transaction.Amount)
				assert.Equal(t, tt.expectedResult.Transaction.Description, response.Transaction.Description)
				assert.Equal(t, tt.expectedResult.Transaction.Status, response.Transaction.Status)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetTransactionHistory(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.GetTransactionHistoryRequest
		mockSetup     func(sqlmock.Sqlmock)
		expectedError string
		expectedTotal int32
		expectedCount int
	}{
		{
			name: "successful transaction history retrieval",
			request: &pb.GetTransactionHistoryRequest{
				AccountId: "test-account-id",
				Limit:     10,
				Offset:    0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM transactions WHERE account_id = \$1`).
					WithArgs("test-account-id").
					WillReturnRows(countRows)

				// Mock transactions query
				rows := sqlmock.NewRows([]string{"id", "account_id", "operation_type", "amount", "description", "created_at", "status"}).
					AddRow("tx1", "test-account-id", "PAYMENT", 100.50, "Payment 1", 1234567890, "COMPLETED").
					AddRow("tx2", "test-account-id", "CASH_PURCHASE", -50.00, "Purchase 1", 1234567891, "COMPLETED")
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("test-account-id", 10, 0).
					WillReturnRows(rows)
			},
			expectedError: "",
			expectedTotal: 2,
			expectedCount: 2,
		},
		{
			name: "missing account id",
			request: &pb.GetTransactionHistoryRequest{
				AccountId: "",
				Limit:     10,
				Offset:    0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "account_id required",
			expectedTotal: 0,
			expectedCount: 0,
		},
		{
			name: "default limit and offset",
			request: &pb.GetTransactionHistoryRequest{
				AccountId: "test-account-id",
				Limit:     0,  // Should default to 50
				Offset:    -1, // Should default to 0
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM transactions WHERE account_id = \$1`).
					WithArgs("test-account-id").
					WillReturnRows(countRows)

				// Mock transactions query with default values
				rows := sqlmock.NewRows([]string{"id", "account_id", "operation_type", "amount", "description", "created_at", "status"})
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("test-account-id", 50, 0).
					WillReturnRows(rows)
			},
			expectedError: "",
			expectedTotal: 0,
			expectedCount: 0,
		},
		{
			name: "limit too high",
			request: &pb.GetTransactionHistoryRequest{
				AccountId: "test-account-id",
				Limit:     150, // Should default to 50 (not 100)
				Offset:    0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock count query
				countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM transactions WHERE account_id = \$1`).
					WithArgs("test-account-id").
					WillReturnRows(countRows)

				// Mock transactions query with default limit (50, not 100)
				rows := sqlmock.NewRows([]string{"id", "account_id", "operation_type", "amount", "description", "created_at", "status"})
				mock.ExpectQuery(`SELECT id, account_id, operation_type, amount, description, created_at, status`).
					WithArgs("test-account-id", 50, 0).
					WillReturnRows(rows)
			},
			expectedError: "",
			expectedTotal: 0,
			expectedCount: 0,
		},
		{
			name: "database error on count",
			request: &pb.GetTransactionHistoryRequest{
				AccountId: "test-account-id",
				Limit:     10,
				Offset:    0,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM transactions WHERE account_id = \$1`).
					WithArgs("test-account-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "database error",
			expectedTotal: 0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			service := NewService(db)
			response, err := service.GetTransactionHistory(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			assert.Equal(t, tt.expectedTotal, response.Total)
			assert.Equal(t, tt.expectedCount, len(response.Transactions))

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_ProcessPayment(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.ProcessPaymentRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.ProcessPaymentResponse
	}{
		{
			name: "successful payment processing",
			request: &pb.ProcessPaymentRequest{
				AccountId:   "test-account-id",
				Amount:      100.50,
				Description: "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 200.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)

				// Mock balance update
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs(100.50, sqlmock.AnyArg(), "test-account-id").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock transaction insert
				mock.ExpectExec(`INSERT INTO transactions`).
					WithArgs(sqlmock.AnyArg(), "test-account-id", "PAYMENT", 100.50, "Test payment", sqlmock.AnyArg(), "COMPLETED").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
			expectedResult: &pb.ProcessPaymentResponse{
				Transaction: &pb.Transaction{
					AccountId:     "test-account-id",
					OperationType: "PAYMENT",
					Amount:        100.50,
					Description:   "Test payment",
					Status:        "COMPLETED",
				},
			},
		},
		{
			name: "error in create transaction",
			request: &pb.ProcessPaymentRequest{
				AccountId:   "test-account-id",
				Amount:      100.50,
				Description: "Test payment",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Mock account lookup
				accountRows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 200.00, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(accountRows)

				// Mock balance update
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs(100.50, sqlmock.AnyArg(), "test-account-id").
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock transaction insert error
				mock.ExpectExec(`INSERT INTO transactions`).
					WithArgs(sqlmock.AnyArg(), "test-account-id", "PAYMENT", 100.50, "Test payment", sqlmock.AnyArg(), "COMPLETED").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "could not create transaction",
			expectedResult: &pb.ProcessPaymentResponse{
				Error: "could not create transaction",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			service := NewService(db)
			response, err := service.ProcessPayment(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedError == "" {
				assert.NotEmpty(t, response.Transaction.Id)
				assert.Equal(t, tt.request.AccountId, response.Transaction.AccountId)
				assert.Equal(t, "PAYMENT", response.Transaction.OperationType)
				assert.Equal(t, tt.request.Amount, response.Transaction.Amount)
				assert.Equal(t, tt.request.Description, response.Transaction.Description)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
