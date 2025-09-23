package health

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthChecker(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	checker := NewHealthChecker(db)
	assert.NotNil(t, checker)
	assert.Equal(t, db, checker.db)
}

func TestHealthChecker_Check(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
	}{
		{
			name: "successful health check",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(nil)
			},
			expectErr: false,
		},
		{
			name: "database ping fails",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			expectErr: true,
		},
		{
			name: "database connection timeout",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			checker := NewHealthChecker(db)
			err = checker.Check(context.Background())

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHealthChecker_CheckWithTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		mockSetup func(sqlmock.Sqlmock)
		expectErr bool
	}{
		{
			name:    "successful health check with timeout",
			timeout: 5 * time.Second,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(nil)
			},
			expectErr: false,
		},
		{
			name:    "health check timeout",
			timeout: 1 * time.Millisecond, // Very short timeout
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Simulate slow database response
				mock.ExpectPing().WillDelayFor(10 * time.Millisecond).WillReturnError(nil)
			},
			expectErr: true,
		},
		{
			name:    "database ping fails with timeout",
			timeout: 5 * time.Second,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			checker := NewHealthChecker(db)
			err = checker.CheckWithTimeout(tt.timeout)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHealthChecker_Check_ContextCancellation(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Don't expect ping since context is cancelled before ping is called
	checker := NewHealthChecker(db)
	err = checker.Check(ctx)

	// Should return context cancelled error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthChecker_Check_ContextTimeout(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Mock ping to take longer than context timeout
	mock.ExpectPing().WillDelayFor(10 * time.Millisecond).WillReturnError(nil)

	checker := NewHealthChecker(db)
	err = checker.Check(ctx)

	// Should return context deadline exceeded error (or similar timeout error)
	assert.Error(t, err)
	// The error message might vary, so just check that it's a timeout-related error
	assert.True(t,
		strings.Contains(err.Error(), "context deadline exceeded") ||
			strings.Contains(err.Error(), "canceling query due to user request") ||
			strings.Contains(err.Error(), "timeout"))

	assert.NoError(t, mock.ExpectationsWereMet())
}
