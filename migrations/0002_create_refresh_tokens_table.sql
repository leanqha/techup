-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
                                id SERIAL PRIMARY KEY,
                                account_id INT REFERENCES accounts(id) ON DELETE CASCADE,
                                token TEXT NOT NULL UNIQUE,
                                expires_at TIMESTAMP NOT NULL,
                                created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;