-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    subject_id INT NULL REFERENCES subjects(id) ON DELETE SET NULL,
    lesson_id INT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    date DATE NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, lesson_id)
);

CREATE INDEX IF NOT EXISTS idx_notes_user_id
    ON notes (user_id);

CREATE INDEX IF NOT EXISTS idx_notes_lesson_id
    ON notes (lesson_id);

CREATE INDEX IF NOT EXISTS idx_notes_subject_id
    ON notes (subject_id);

CREATE INDEX IF NOT EXISTS idx_notes_date
    ON notes (date);

-- +goose Down
DROP TABLE IF EXISTS notes;
