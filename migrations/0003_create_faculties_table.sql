-- +goose Up
CREATE TABLE faculties (
                           id SERIAL PRIMARY KEY,
                           name TEXT NOT NULL,
                           type TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS faculties CASCADE ;