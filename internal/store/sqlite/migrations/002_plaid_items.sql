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
