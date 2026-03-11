CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    plaid_transaction_id TEXT NOT NULL UNIQUE,
    account_id INTEGER NOT NULL REFERENCES account (id),
    amount REAL NOT NULL,
    transaction_date TEXT NOT NULL,
    description TEXT NOT NULL,
    merchant_name TEXT,
    pending INTEGER NOT NULL,
    payment_channel TEXT NOT NULL,
    plaid_category_primary TEXT,
    plaid_category_detailed TEXT,
    category_id INTEGER REFERENCES category (id),
    last_modified_by INTEGER REFERENCES user (id),
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_transactions_account_id
ON transactions (account_id);

CREATE INDEX IF NOT EXISTS idx_transactions_date
ON transactions (transaction_date);
