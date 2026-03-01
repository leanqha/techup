-- +goose Up
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_expires
    ON password_reset_tokens (expires_at);

CREATE INDEX idx_password_reset_tokens_account
    ON password_reset_tokens (account_id);

-- +goose Down
DROP TABLE IF EXISTS password_reset_tokens;

