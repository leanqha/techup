-- +goose Up
CREATE TABLE IF NOT EXISTS buildings (
                                         id INT PRIMARY KEY,
                                         name TEXT NOT NULL,
                                         address TEXT NOT NULL,
                                         created_at TIMESTAMP DEFAULT NOW(),
                                         updated_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO buildings (id, name, address)
VALUES (0, 'Тестовый корпус', 'Тестовая улица 1')
ON CONFLICT DO NOTHING;

INSERT INTO buildings (id, name, address)
VALUES (1, 'Главный корпус', 'Московский проспект 24')
ON CONFLICT DO NOTHING;

INSERT INTO buildings (id, name, address)
VALUES (2, 'Корпус на 7-й Красноармейской', '7-я Красноармейская улица 6')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS buildings CASCADE;