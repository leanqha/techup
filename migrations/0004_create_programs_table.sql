-- +goose Up
CREATE TABLE programs (
                          id SERIAL PRIMARY KEY,
                          faculty_id INT REFERENCES faculties(id),
                          name TEXT NOT NULL,
                          degree TEXT NOT NULL,
                          course INT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS programs CASCADE ;