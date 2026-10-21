-- +goose Up
-- +goose StatementBegin
-- Each curriculum entry can be tied to a lesson type (Лекция, Практика, ...)
-- from the lesson_types catalog. Optional: dropping a type unsets it rather
-- than deleting the plan entry.
ALTER TABLE group_subjects
    ADD COLUMN IF NOT EXISTS lesson_type_id INT REFERENCES lesson_types(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_subjects DROP COLUMN IF EXISTS lesson_type_id;
-- +goose StatementEnd
