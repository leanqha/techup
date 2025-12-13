-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
                                        id SERIAL PRIMARY KEY,
                                        uid TEXT DEFAULT '000000',
                                        email TEXT UNIQUE NOT NULL,
                                        password_hash TEXT NOT NULL,
                                        first_name TEXT,
                                        last_name TEXT,
                                        role TEXT DEFAULT 'student',
                                        is_verified BOOLEAN NOT NULL DEFAULT false,
                                        group_id INT REFERENCES groups(id),
                                        created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                        updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS accounts CASCADE;