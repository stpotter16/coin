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

CREATE TABLE IF NOT EXISTS account (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;

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

-- All transactions are manually entered by the user.
-- Assigned to a plan item or left unassigned (flexible spending).
CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY,
    account_id INTEGER REFERENCES account (id),
    plan_item_id INTEGER REFERENCES plan_items (id),
    amount REAL NOT NULL,
    transaction_date TEXT NOT NULL,
    description TEXT NOT NULL,
    merchant_name TEXT,
    pending INTEGER NOT NULL DEFAULT 0,
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
