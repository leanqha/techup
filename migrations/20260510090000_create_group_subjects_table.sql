-- +goose Up
-- +goose StatementBegin
-- group_subjects is a group's curriculum for one semester: which subjects the
-- group studies in a given semester, and (optionally) who teaches each. It is a
-- separate catalog from lessons — lessons are concrete dated pairs, while this
-- is the planned subject list per semester.
CREATE TABLE IF NOT EXISTS group_subjects (
    id SERIAL PRIMARY KEY,
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    subject_id INT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    semester INT NOT NULL CHECK (semester > 0),
    teacher_id INT REFERENCES accounts(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A subject appears at most once per group per semester.
CREATE UNIQUE INDEX IF NOT EXISTS idx_group_subjects_unique
    ON group_subjects (group_id, subject_id, semester);

-- Listing a group's plan for a semester is the common read path.
CREATE INDEX IF NOT EXISTS idx_group_subjects_group_semester
    ON group_subjects (group_id, semester);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_subjects;
-- +goose StatementEnd
