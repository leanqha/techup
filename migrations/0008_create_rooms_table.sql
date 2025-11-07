-- +goose Up
CREATE TABLE IF NOT EXISTS rooms (
                                     id SERIAL PRIMARY KEY,
                                     building_id INT REFERENCES buildings(id) ON DELETE CASCADE,
                                     floor INT NOT NULL,
                                     name TEXT UNIQUE NOT NULL,
                                     created_at TIMESTAMP DEFAULT NOW(),
                                     updated_at TIMESTAMP DEFAULT NOW()
);

-- вставка кабинетов для "Тестового корпуса" (id = 0) без циклов
INSERT INTO rooms (building_id, floor, name) VALUES
                                                 (0, 1, '101'),
                                                 (0, 1, '102'),
                                                 (0, 1, '103'),
                                                 (0, 1, '104'),
                                                 (0, 1, '105'),
                                                 (0, 1, '106'),
                                                 (0, 1, '107'),
                                                 (0, 1, '108'),
                                                 (0, 1, '109'),
                                                 (0, 1, '110'),
                                                 (0, 1, '111'),
                                                 (0, 1, '112'),
                                                 (0, 1, '113'),
                                                 (0, 1, '114'),
                                                 (0, 1, '115'),
                                                 (0, 2, '201'),
                                                 (0, 2, '202'),
                                                 (0, 2, '203'),
                                                 (0, 2, '204'),
                                                 (0, 2, '205'),
                                                 (0, 2, '206'),
                                                 (0, 2, '207'),
                                                 (0, 2, '208'),
                                                 (0, 2, '209'),
                                                 (0, 2, '210'),
                                                 (0, 2, '211'),
                                                 (0, 2, '212'),
                                                 (0, 2, '213'),
                                                 (0, 2, '214'),
                                                 (0, 2, '215');

-- +goose Down
DROP TABLE IF EXISTS rooms CASCADE;