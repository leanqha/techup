-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS lesson_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#64748b',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lesson_types_name_unique
    ON lesson_types (LOWER(name));

-- Seed the initial catalog. Colors double as the lesson card accent on the
-- frontend; new types (Экзамен, Зачет) get distinct hues, and "Другое" covers
-- the lessons.type default of 'other'.
INSERT INTO lesson_types (name, color) VALUES
    ('Лекция', '#2563eb'),
    ('Практика', '#dc2626'),
    ('Лабораторная', '#7c3aed'),
    ('Экзамен', '#ea580c'),
    ('Зачет', '#059669'),
    ('Другое', '#64748b')
ON CONFLICT DO NOTHING;

-- Lessons stored the type as an English code; move them to the catalog's
-- display names so lessons.type matches a lesson_types.name.
UPDATE lessons SET type = 'Лекция'       WHERE type = 'lecture';
UPDATE lessons SET type = 'Практика'     WHERE type = 'practice';
UPDATE lessons SET type = 'Лабораторная' WHERE type = 'laboratory';
UPDATE lessons SET type = 'Экзамен'      WHERE type = 'exam';
UPDATE lessons SET type = 'Зачет'        WHERE type IN ('credit', 'zachet');
UPDATE lessons SET type = 'Другое'       WHERE type = 'other';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert the known display names back to their English codes. Custom types
-- added through the admin keep their names (there is no code to restore).
UPDATE lessons SET type = 'lecture'    WHERE type = 'Лекция';
UPDATE lessons SET type = 'practice'   WHERE type = 'Практика';
UPDATE lessons SET type = 'laboratory' WHERE type = 'Лабораторная';
UPDATE lessons SET type = 'exam'       WHERE type = 'Экзамен';
UPDATE lessons SET type = 'credit'     WHERE type = 'Зачет';
UPDATE lessons SET type = 'other'      WHERE type = 'Другое';

DROP TABLE IF EXISTS lesson_types;
-- +goose StatementEnd
