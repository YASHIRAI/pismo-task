-- PostgreSQL Database initialization script for Pismo microservices

-- Create accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(36) PRIMARY KEY,
    document_number VARCHAR(20) NOT NULL UNIQUE,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('CHECKING', 'SAVINGS', 'CREDIT')),
    balance DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(36) NOT NULL,
    operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN ('CASH_PURCHASE', 'INSTALLMENT_PURCHASE', 'WITHDRAWAL', 'PAYMENT')),
    amount DECIMAL(15,2) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_accounts_document_number ON accounts(document_number);
CREATE INDEX IF NOT EXISTS idx_accounts_account_type ON accounts(account_type);
CREATE INDEX IF NOT EXISTS idx_accounts_created_at ON accounts(created_at);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_account_created ON transactions(account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_operation_type ON transactions(operation_type);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- Insert test data
INSERT INTO accounts (id, document_number, account_type, balance, created_at, updated_at) 
VALUES (
    'test-account-123', 
    '12345678901', 
    'CHECKING', 
    1000.00, 
    EXTRACT(EPOCH FROM NOW()), 
    EXTRACT(EPOCH FROM NOW())
) ON CONFLICT (id) DO NOTHING;

-- Insert test transaction
INSERT INTO transactions (id, account_id, operation_type, amount, description, created_at, status)
VALUES (
    'test-transaction-123',
    'test-account-123',
    'PAYMENT',
    100.00,
    'Test payment transaction',
    EXTRACT(EPOCH FROM NOW()),
    'COMPLETED'
) ON CONFLICT (id) DO NOTHING;