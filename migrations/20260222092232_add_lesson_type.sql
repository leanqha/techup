-- +goose Up
ALTER TABLE lessons
    ADD COLUMN type TEXT NOT NULL DEFAULT 'other';

-- +goose Down
ALTER TABLE lessons
    DROP COLUMN type;