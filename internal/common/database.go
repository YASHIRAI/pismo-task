package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type DatabaseManager struct {
	db     *sql.DB
	config DatabaseConfig
}

func NewDatabaseManager() (*DatabaseManager, error) {
	config := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "pismo"),
		Password: getEnv("DB_PASSWORD", "pismo123"),
		DBName:   getEnv("DB_NAME", "pismo"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseManager{
		db:     db,
		config: config,
	}, nil
}

func (dm *DatabaseManager) GetDB() *sql.DB {
	return dm.db
}

func (dm *DatabaseManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}

func (dm *DatabaseManager) Health() error {
	return dm.db.Ping()
}

func (dm *DatabaseManager) InitSchema() error {
	_, err := dm.db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id VARCHAR(36) PRIMARY KEY,
			document_number VARCHAR(20) NOT NULL UNIQUE,
			account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('CHECKING', 'SAVINGS', 'CREDIT')),
			balance DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
			created_at BIGINT NOT NULL,
			updated_at BIGINT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create accounts table: %w", err)
	}

	_, err = dm.db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id VARCHAR(36) PRIMARY KEY,
			account_id VARCHAR(36) NOT NULL,
			operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN ('CASH_PURCHASE', 'INSTALLMENT_PURCHASE', 'WITHDRAWAL', 'PAYMENT')),
			amount DECIMAL(15,2) NOT NULL,
			description TEXT,
			created_at BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_accounts_document_number ON accounts(document_number)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_account_type ON accounts(account_type)",
		"CREATE INDEX IF NOT EXISTS idx_accounts_created_at ON accounts(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_account_created ON transactions(account_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_operation_type ON transactions(operation_type)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status)",
	}

	for _, indexSQL := range indexes {
		if _, err := dm.db.Exec(indexSQL); err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
