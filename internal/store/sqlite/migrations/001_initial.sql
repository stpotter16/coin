CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_admin INTEGER NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS session (
    session_key TEXT PRIMARY KEY,
    value BLOB,
    expires_at TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS plaid_item (
    id INTEGER PRIMARY KEY,
    plaid_item_id TEXT NOT NULL UNIQUE,
    plaid_access_token TEXT NOT NULL,
    institution_id TEXT NOT NULL,
    institution_name TEXT NOT NULL,
    transaction_cursor TEXT,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

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

-- Raw Plaid transaction cache. One row per Plaid transaction ID.
-- Append-only from the sync job; never modified by user actions.
CREATE TABLE IF NOT EXISTS plaid_transactions (
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
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_plaid_transactions_account_id
ON plaid_transactions (account_id);

CREATE INDEX IF NOT EXISTS idx_plaid_transactions_date
ON plaid_transactions (transaction_date);

-- Plans: one per user-month. Locked plans cannot be edited.
CREATE TABLE IF NOT EXISTS plans (
    id INTEGER PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    locked INTEGER NOT NULL DEFAULT 0,
    created_by INTEGER NOT NULL REFERENCES user (id),
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL,
    UNIQUE (year, month, created_by)
) STRICT;

-- Plan items: income sources or fixed expenses within a plan.
CREATE TABLE IF NOT EXISTS plan_items (
    id INTEGER PRIMARY KEY,
    plan_id INTEGER NOT NULL REFERENCES plans (id),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('income', 'fixed_expense')),
    expected_amount REAL NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_plan_items_plan_id ON plan_items (plan_id);

-- Domain transactions. May be sourced from Plaid or entered manually.
-- Assigned to a plan item or left unassigned (flexible spending).
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    plaid_transaction_id INTEGER REFERENCES plaid_transactions (id),
    account_id INTEGER REFERENCES account (id),
    plan_item_id INTEGER REFERENCES plan_items (id),
    amount REAL NOT NULL,
    transaction_date TEXT NOT NULL,
    description TEXT NOT NULL,
    merchant_name TEXT,
    pending INTEGER NOT NULL DEFAULT 0,
    payment_channel TEXT,
    plaid_category_primary TEXT,
    plaid_category_detailed TEXT,
    created_by INTEGER REFERENCES user (id),
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_transactions_account_id
ON transactions (account_id);

CREATE INDEX IF NOT EXISTS idx_transactions_date
ON transactions (transaction_date);

CREATE INDEX IF NOT EXISTS idx_transactions_plan_item_id
ON transactions (plan_item_id);
