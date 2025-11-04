-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
                                        id SERIAL PRIMARY KEY,
                                        email TEXT UNIQUE NOT NULL,
                                        password_hash TEXT NOT NULL,
                                        first_name TEXT,
                                        last_name TEXT,
                                        role TEXT DEFAULT 'student',
                                        created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );

-- +goose Down
DROP TABLE IF EXISTS accounts;