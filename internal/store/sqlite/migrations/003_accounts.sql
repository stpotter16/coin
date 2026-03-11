CREATE TABLE IF NOT EXISTS account (
    id INTEGER PRIMARY KEY,
    plaid_account_id TEXT NOT NULL UNIQUE,
    plaid_item_id INTEGER NOT NULL REFERENCES plaid_item (id),
    name TEXT NOT NULL,
    official_name TEXT,
    type TEXT NOT NULL,
    subtype TEXT NOT NULL,
    current_balance REAL,
    available_balance REAL,
    iso_currency_code TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_account_plaid_item_id ON account (plaid_item_id);
