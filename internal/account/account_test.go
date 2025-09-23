package account

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/YASHIRAI/pismo-task/internal/common"
	pb "github.com/YASHIRAI/pismo-task/proto/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	logger, _ := common.NewLogger("test-service", common.INFO)
	service := NewService(db, logger)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestService_CreateAccount(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.CreateAccountRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.CreateAccountResponse
	}{
		{
			name: "successful account creation",
			request: &pb.CreateAccountRequest{
				DocumentNumber: "12345678901",
				AccountType:    "CHECKING",
				InitialBalance: 100.50,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO accounts`).
					WithArgs(sqlmock.AnyArg(), "12345678901", "CHECKING", 100.50, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
			expectedResult: &pb.CreateAccountResponse{
				Account: &pb.Account{
					DocumentNumber: "12345678901",
					AccountType:    "CHECKING",
					Balance:        100.50,
				},
			},
		},
		{
			name: "missing document number",
			request: &pb.CreateAccountRequest{
				DocumentNumber: "",
				AccountType:    "CHECKING",
				InitialBalance: 100.50,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "missing required fields",
			expectedResult: &pb.CreateAccountResponse{
				Error: "missing required fields",
			},
		},
		{
			name: "missing account type",
			request: &pb.CreateAccountRequest{
				DocumentNumber: "12345678901",
				AccountType:    "",
				InitialBalance: 100.50,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "missing required fields",
			expectedResult: &pb.CreateAccountResponse{
				Error: "missing required fields",
			},
		},
		{
			name: "database error",
			request: &pb.CreateAccountRequest{
				DocumentNumber: "12345678901",
				AccountType:    "CHECKING",
				InitialBalance: 100.50,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO accounts`).
					WithArgs(sqlmock.AnyArg(), "12345678901", "CHECKING", 100.50, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "could not create account",
			expectedResult: &pb.CreateAccountResponse{
				Error: "could not create account",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			logger, _ := common.NewLogger("test-service", common.INFO)
			service := NewService(db, logger)
			response, err := service.CreateAccount(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedError == "" {
				assert.NotEmpty(t, response.Account.Id)
				assert.Equal(t, tt.request.DocumentNumber, response.Account.DocumentNumber)
				assert.Equal(t, tt.request.AccountType, response.Account.AccountType)
				assert.Equal(t, tt.request.InitialBalance, response.Account.Balance)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetAccount(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.GetAccountRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.GetAccountResponse
	}{
		{
			name: "successful account retrieval",
			request: &pb.GetAccountRequest{
				Id: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "12345678901", "CHECKING", 100.50, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(rows)
			},
			expectedError: "",
			expectedResult: &pb.GetAccountResponse{
				Account: &pb.Account{
					Id:             "test-account-id",
					DocumentNumber: "12345678901",
					AccountType:    "CHECKING",
					Balance:        100.50,
					CreatedAt:      1234567890,
					UpdatedAt:      1234567890,
				},
			},
		},
		{
			name: "missing account id",
			request: &pb.GetAccountRequest{
				Id: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "id required",
			expectedResult: &pb.GetAccountResponse{
				Error: "id required",
			},
		},
		{
			name: "account not found",
			request: &pb.GetAccountRequest{
				Id: "non-existent-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("non-existent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: "not found",
			expectedResult: &pb.GetAccountResponse{
				Error: "not found",
			},
		},
		{
			name: "database error",
			request: &pb.GetAccountRequest{
				Id: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "database error",
			expectedResult: &pb.GetAccountResponse{
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

			logger, _ := common.NewLogger("test-service", common.INFO)
			service := NewService(db, logger)
			response, err := service.GetAccount(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedError == "" {
				assert.Equal(t, tt.expectedResult.Account.Id, response.Account.Id)
				assert.Equal(t, tt.expectedResult.Account.DocumentNumber, response.Account.DocumentNumber)
				assert.Equal(t, tt.expectedResult.Account.AccountType, response.Account.AccountType)
				assert.Equal(t, tt.expectedResult.Account.Balance, response.Account.Balance)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_UpdateAccount(t *testing.T) {
	tests := []struct {
		name          string
		request       *pb.UpdateAccountRequest
		mockSetup     func(sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name: "successful account update",
			request: &pb.UpdateAccountRequest{
				Id:             "test-account-id",
				DocumentNumber: "98765432109",
				AccountType:    "SAVINGS",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs("test-account-id", "98765432109", "SAVINGS", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Mock the GetAccount call that happens after update
				rows := sqlmock.NewRows([]string{"id", "document_number", "account_type", "balance", "created_at", "updated_at"}).
					AddRow("test-account-id", "98765432109", "SAVINGS", 100.50, 1234567890, 1234567890)
				mock.ExpectQuery(`SELECT id, document_number, account_type, balance, created_at, updated_at`).
					WithArgs("test-account-id").
					WillReturnRows(rows)
			},
			expectedError: "",
		},
		{
			name: "missing account id",
			request: &pb.UpdateAccountRequest{
				Id: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "id required",
		},
		{
			name: "database error on update",
			request: &pb.UpdateAccountRequest{
				Id:             "test-account-id",
				DocumentNumber: "98765432109",
				AccountType:    "SAVINGS",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE accounts`).
					WithArgs("test-account-id", "98765432109", "SAVINGS", sqlmock.AnyArg()).
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "could not update account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			logger, _ := common.NewLogger("test-service", common.INFO)
			service := NewService(db, logger)
			response, err := service.UpdateAccount(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_DeleteAccount(t *testing.T) {
	tests := []struct {
		name           string
		request        *pb.DeleteAccountRequest
		mockSetup      func(sqlmock.Sqlmock)
		expectedError  string
		expectedResult *pb.DeleteAccountResponse
	}{
		{
			name: "successful account deletion",
			request: &pb.DeleteAccountRequest{
				Id: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM accounts WHERE id = \$1`).
					WithArgs("test-account-id").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: "",
			expectedResult: &pb.DeleteAccountResponse{
				Success: true,
			},
		},
		{
			name: "missing account id",
			request: &pb.DeleteAccountRequest{
				Id: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError: "id required",
			expectedResult: &pb.DeleteAccountResponse{
				Error: "id required",
			},
		},
		{
			name: "account not found",
			request: &pb.DeleteAccountRequest{
				Id: "non-existent-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM accounts WHERE id = \$1`).
					WithArgs("non-existent-id").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: "account not found",
			expectedResult: &pb.DeleteAccountResponse{
				Error: "account not found",
			},
		},
		{
			name: "database error",
			request: &pb.DeleteAccountRequest{
				Id: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM accounts WHERE id = \$1`).
					WithArgs("test-account-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: "could not delete account",
			expectedResult: &pb.DeleteAccountResponse{
				Error: "could not delete account",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			logger, _ := common.NewLogger("test-service", common.INFO)
			service := NewService(db, logger)
			response, err := service.DeleteAccount(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			if tt.expectedResult != nil {
				assert.Equal(t, tt.expectedResult.Success, response.Success)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestService_GetBalance(t *testing.T) {
	tests := []struct {
		name            string
		request         *pb.GetBalanceRequest
		mockSetup       func(sqlmock.Sqlmock)
		expectedError   string
		expectedBalance float64
	}{
		{
			name: "successful balance retrieval",
			request: &pb.GetBalanceRequest{
				AccountId: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"balance"}).
					AddRow(150.75)
				mock.ExpectQuery(`SELECT balance FROM accounts WHERE id = \$1`).
					WithArgs("test-account-id").
					WillReturnRows(rows)
			},
			expectedError:   "",
			expectedBalance: 150.75,
		},
		{
			name: "missing account id",
			request: &pb.GetBalanceRequest{
				AccountId: "",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// No database call expected
			},
			expectedError:   "account_id required",
			expectedBalance: 0,
		},
		{
			name: "account not found",
			request: &pb.GetBalanceRequest{
				AccountId: "non-existent-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT balance FROM accounts WHERE id = \$1`).
					WithArgs("non-existent-id").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:   "account not found",
			expectedBalance: 0,
		},
		{
			name: "database error",
			request: &pb.GetBalanceRequest{
				AccountId: "test-account-id",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT balance FROM accounts WHERE id = \$1`).
					WithArgs("test-account-id").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError:   "database error",
			expectedBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			logger, _ := common.NewLogger("test-service", common.INFO)
			service := NewService(db, logger)
			response, err := service.GetBalance(context.Background(), tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, response.Error)
			assert.Equal(t, tt.expectedBalance, response.Balance)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
