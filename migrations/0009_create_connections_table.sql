-- +goose Up
CREATE TABLE IF NOT EXISTS connections (
                             id SERIAL PRIMARY KEY,
                             room_from TEXT REFERENCES rooms(name),
                             room_to TEXT REFERENCES rooms(name),
                             distance FLOAT NOT NULL,
                             type TEXT DEFAULT 'corridor',   -- corridor / stairs
                             created_at TIMESTAMP DEFAULT NOW(),
                             updated_at TIMESTAMP DEFAULT NOW()
);

-- Test data: Connections for test building (building_id=3)
-- Rooms: 101–115 (first floor), 201–215 (second floor)
-- Corridor connections (consecutive rooms)
INSERT INTO connections (room_from, room_to, distance, type)
VALUES
    (101, 102, 1.0, 'corridor'),
    (102, 103, 1.0, 'corridor'),
    (103, 104, 1.0, 'corridor'),
    (104, 105, 1.0, 'corridor'),
    (105, 106, 1.0, 'corridor'),
    (106, 107, 1.0, 'corridor'),
    (107, 108, 1.0, 'corridor'),
    (108, 109, 1.0, 'corridor'),
    (109, 110, 1.0, 'corridor'),
    (110, 111, 1.0, 'corridor'),
    (111, 112, 1.0, 'corridor'),
    (112, 113, 1.0, 'corridor'),
    (113, 114, 1.0, 'corridor'),
    (114, 115, 1.0, 'corridor'),
    (201, 202, 1.0, 'corridor'),
    (202, 203, 1.0, 'corridor'),
    (203, 204, 1.0, 'corridor'),
    (204, 205, 1.0, 'corridor'),
    (205, 206, 1.0, 'corridor'),
    (206, 207, 1.0, 'corridor'),
    (207, 208, 1.0, 'corridor'),
    (208, 209, 1.0, 'corridor'),
    (209, 210, 1.0, 'corridor'),
    (210, 211, 1.0, 'corridor'),
    (211, 212, 1.0, 'corridor'),
    (212, 213, 1.0, 'corridor'),
    (213, 214, 1.0, 'corridor'),
    (214, 215, 1.0, 'corridor');

-- Main stairs connecting first and second floor
INSERT INTO connections (room_from, room_to, distance, type)
VALUES (101, 201, 1.0, 'stairs');

-- +goose Down
DROP TABLE IF EXISTS connections CASCADE;