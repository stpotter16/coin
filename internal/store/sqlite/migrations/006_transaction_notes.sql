CREATE TABLE IF NOT EXISTS transaction_notes (
    id INTEGER PRIMARY KEY,
    transaction_id INTEGER NOT NULL REFERENCES transactions (id),
    user_id INTEGER NOT NULL REFERENCES user (id),
    note TEXT NOT NULL,
    created_time TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_transaction_notes_transaction_id
ON transaction_notes (transaction_id);
