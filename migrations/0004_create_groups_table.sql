-- +goose Up
CREATE TABLE IF NOT EXISTS groups (
                        id SERIAL PRIMARY KEY,
                        faculty_id INT REFERENCES faculties(id) ON DELETE CASCADE,
                        name TEXT NOT NULL,                        course INT NOT NULL,             -- курс
                        degree TEXT NOT NULL,            -- бакалавриат / магистратура
                        year_start INT NOT NULL,         -- год набора
                        specialization TEXT,             -- профиль / направление
                        is_active BOOLEAN DEFAULT true
);

-- +goose Down
DROP TABLE IF EXISTS groups CASCADE ;