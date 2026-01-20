-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id            BIGSERIAL PRIMARY KEY,
    amount        DECIMAL(15,2) NOT NULL,
    type          VARCHAR(20) NOT NULL CHECK (type IN ('in', 'out')),
    category      VARCHAR(50),
    description   TEXT,
    source        VARCHAR(100) NOT NULL,
    source_account VARCHAR(100),
    recipient     VARCHAR(100),
    transaction_date TIMESTAMP NOT NULL,
    created_at    TIMESTAMP DEFAULT NOW(),
    updated_at    TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_source ON transactions(source);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);

-- Create a comment for documentation
COMMENT ON TABLE transactions IS 'Stores financial transactions from various banks and e-wallets';
COMMENT ON COLUMN transactions.type IS 'Transaction type: "in" for income, "out" for expense';
COMMENT ON COLUMN transactions.source IS 'Bank or e-wallet name where the transaction occurred';
COMMENT ON COLUMN transactions.source_account IS 'Account identifier (e.g., account number, wallet ID)';
