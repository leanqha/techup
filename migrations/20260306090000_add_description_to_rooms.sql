-- +goose Up
-- The `rooms` table is created by a later migration; on a fresh database it
-- does not exist yet, so only alter it when present (idempotent either way).
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.rooms') IS NOT NULL THEN
        ALTER TABLE rooms ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF to_regclass('public.rooms') IS NOT NULL THEN
        ALTER TABLE rooms DROP COLUMN IF EXISTS description;
    END IF;
END $$;
-- +goose StatementEnd
