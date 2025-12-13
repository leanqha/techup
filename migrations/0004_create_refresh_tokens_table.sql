-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
                                id SERIAL PRIMARY KEY,
                                account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
                                token TEXT NOT NULL UNIQUE,
                                expires_at TIMESTAMP NOT NULL,
                                created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_expires
    ON refresh_tokens (expires_at);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;