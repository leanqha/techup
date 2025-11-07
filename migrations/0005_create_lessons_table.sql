-- +goose Up
CREATE TABLE IF NOT EXISTS lessons (
                         id SERIAL PRIMARY KEY,
                         group_id INT REFERENCES groups(id) ON DELETE CASCADE,
                         group_name TEXT NOT NULL,
                         day_of_week TEXT NOT NULL,
                         start_time TIME NOT NULL,
                         end_time TIME NOT NULL,
                         subject TEXT NOT NULL,
                         teacher TEXT,
                         classroom TEXT,
                         is_online BOOLEAN DEFAULT false,
                         is_even_week BOOLEAN DEFAULT true,
                         created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS lessons;