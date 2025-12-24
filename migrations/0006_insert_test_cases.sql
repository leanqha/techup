-- +goose Up
-- Создание тестового факультета и группы
INSERT INTO faculties (id, name)
VALUES (0, 'Тестовый факультет')
ON CONFLICT (id) DO NOTHING;

INSERT INTO groups (id, faculty_id, name, course, degree, year_start, specialization, is_active)
VALUES (0, 0, 'Нет', 0, 'Нет', 1970, 'Нет', true)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
-- Удаление тестовых данных
DELETE FROM groups WHERE id = 0;
DELETE FROM faculties WHERE id = 0;