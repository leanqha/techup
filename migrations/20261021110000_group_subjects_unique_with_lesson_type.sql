-- +goose Up
-- +goose StatementBegin
-- A subject may now appear more than once in a group's semester plan as long as
-- the lesson type differs (e.g. Лекция and Практика). Widen the uniqueness to
-- include lesson_type_id. COALESCE keeps two entries with no lesson type from
-- colliding by treating a missing type as a single value.
DROP INDEX IF EXISTS idx_group_subjects_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_subjects_unique
    ON group_subjects (group_id, subject_id, semester, COALESCE(lesson_type_id, 0));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_group_subjects_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_subjects_unique
    ON group_subjects (group_id, subject_id, semester);
-- +goose StatementEnd
