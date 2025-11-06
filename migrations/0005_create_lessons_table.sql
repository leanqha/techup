-- +goose Up
CREATE TABLE lessons (
                         id SERIAL PRIMARY KEY,
                         program_id INT REFERENCES programs(id),
                         day_of_week TEXT,
                         start_time TEXT,
                         end_time TEXT,
                         subject TEXT,
                         teacher TEXT,
                         classroom TEXT,
                         is_online BOOLEAN DEFAULT false,
                         group_number TEXT,
                         created_at TIMESTAMP DEFAULT NOW(),
                         is_even_week BOOLEAN
);

-- +goose Down
DROP TABLE IF EXISTS lessons;