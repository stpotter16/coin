CREATE TABLE IF NOT EXISTS category (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_by INTEGER NOT NULL REFERENCES user (id),
    last_modified_by INTEGER NOT NULL REFERENCES user (id),
    created_time TEXT NOT NULL,
    last_modified_time TEXT NOT NULL
) STRICT;
