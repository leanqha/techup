-- +goose Up
-- +goose StatementBegin
-- Экзаменационные сессии факультета: период с датами начала и конца. У одного
-- факультета может быть несколько сессий (зимняя, летняя и т.д.).
CREATE TABLE IF NOT EXISTS faculty_sessions (
    id SERIAL PRIMARY KEY,
    faculty_id INT NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_faculty_sessions_faculty
    ON faculty_sessions (faculty_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS faculty_sessions;
-- +goose StatementEnd
