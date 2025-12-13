-- +goose Up
-- +goose StatementBegin

CREATE TABLE lessons (
                         id SERIAL PRIMARY KEY,

                         group_id INT NOT NULL
                             REFERENCES groups(id)
                                 ON DELETE CASCADE,

                         date DATE NOT NULL,

                         start_time TIME NOT NULL,
                         end_time TIME NOT NULL,

                         subject TEXT NOT NULL,
                         teacher TEXT NOT NULL,
                         classroom TEXT NOT NULL,

                         created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lessons_group_date
    ON lessons (group_id, date);

CREATE INDEX idx_lessons_date
    ON lessons (date);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lessons;

DROP INDEX IF EXISTS idx_lessons_group_date;
DROP INDEX IF EXISTS idx_lessons_date;

-- +goose StatementEnd