-- +goose Up
-- +goose StatementBegin

CREATE TABLE lesson_notes (
                              id SERIAL PRIMARY KEY,

                              user_id INT NOT NULL,
                              lesson_id INT NOT NULL
                                  REFERENCES lessons(id)
                                      ON DELETE CASCADE,

                              text TEXT NOT NULL,

                              created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                              updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                              UNIQUE (user_id, lesson_id)
);

CREATE INDEX idx_lesson_notes_user
    ON lesson_notes (user_id);

CREATE INDEX idx_lesson_notes_lesson
    ON lesson_notes (lesson_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lesson_notes;

DROP INDEX IF EXISTS idx_lesson_notes_user;
DROP INDEX IF EXISTS idx_lesson_notes_lesson;

-- +goose StatementEnd