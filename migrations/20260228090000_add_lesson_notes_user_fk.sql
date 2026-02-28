-- +goose Up
-- +goose StatementBegin

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'lesson_notes_user_id_fkey'
    ) THEN
        ALTER TABLE lesson_notes
            ADD CONSTRAINT lesson_notes_user_id_fkey
                FOREIGN KEY (user_id)
                REFERENCES accounts(id)
                ON DELETE CASCADE;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'lesson_notes_user_id_fkey'
    ) THEN
        ALTER TABLE lesson_notes
            DROP CONSTRAINT lesson_notes_user_id_fkey;
    END IF;
END $$;

-- +goose StatementEnd

