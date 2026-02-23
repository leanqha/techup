-- +goose Up
CREATE TABLE IF NOT EXISTS faculties (
                                         id SERIAL PRIMARY KEY,
                                         name TEXT NOT NULL
);

-- Insert default faculties if they don't exist
INSERT INTO faculties (id, name)
SELECT 0, 'Тестовый факультет' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=0);
INSERT INTO faculties (id, name)
SELECT 1, 'Химии веществ и материалов' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=1);
INSERT INTO faculties (id, name)
SELECT 2, 'Химической и биотехнологии' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=2);
INSERT INTO faculties (id, name)
SELECT 3, 'Механический' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=3);
INSERT INTO faculties (id, name)
SELECT 4, 'Информационных технологий и управления' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=4);
INSERT INTO faculties (id, name)
SELECT 5, 'Инженерно-технологический' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=5);
INSERT INTO faculties (id, name)
SELECT 6, 'Экономики и менеджмента' WHERE NOT EXISTS (SELECT 1 FROM faculties WHERE id=6);

-- +goose Down
DROP TABLE IF EXISTS faculties CASCADE;
