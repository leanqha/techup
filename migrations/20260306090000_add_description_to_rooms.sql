-- +goose Up
ALTER TABLE rooms
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE rooms
    DROP COLUMN IF EXISTS description;

