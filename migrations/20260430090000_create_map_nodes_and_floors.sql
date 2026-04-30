-- +goose Up
CREATE TABLE IF NOT EXISTS buildings (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL
);

ALTER TABLE buildings
    ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS floors (
    id SERIAL PRIMARY KEY,
    building_id INT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    number INT NOT NULL
);

CREATE TABLE IF NOT EXISTS nodes (
    id SERIAL PRIMARY KEY,
    building_id INT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    floor_id INT NOT NULL REFERENCES floors(id) ON DELETE CASCADE,
    x INT NOT NULL,
    y INT NOT NULL,
    type TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    building_id INT NOT NULL REFERENCES buildings(id) ON DELETE CASCADE,
    floor_id INT NOT NULL REFERENCES floors(id) ON DELETE CASCADE,
    door_node_id INT REFERENCES nodes(id) ON DELETE SET NULL
);

ALTER TABLE rooms
    ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';

ALTER TABLE rooms
    ADD COLUMN IF NOT EXISTS floor_id INT REFERENCES floors(id) ON DELETE CASCADE;

ALTER TABLE rooms
    ADD COLUMN IF NOT EXISTS door_node_id INT REFERENCES nodes(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS connections (
    id SERIAL PRIMARY KEY,
    "from" INT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    "to" INT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    distance FLOAT NOT NULL
);

ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS "from" INT REFERENCES nodes(id) ON DELETE CASCADE;

ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS "to" INT REFERENCES nodes(id) ON DELETE CASCADE;

DELETE FROM connections
WHERE "from" IS NULL OR "to" IS NULL;

ALTER TABLE connections
    ALTER COLUMN "from" SET NOT NULL;

ALTER TABLE connections
    ALTER COLUMN "to" SET NOT NULL;

ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS distance FLOAT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_floors_building_id ON floors(building_id);
CREATE INDEX IF NOT EXISTS idx_nodes_building_id ON nodes(building_id);
CREATE INDEX IF NOT EXISTS idx_nodes_floor_id ON nodes(floor_id);
CREATE INDEX IF NOT EXISTS idx_rooms_building_id ON rooms(building_id);
CREATE INDEX IF NOT EXISTS idx_rooms_floor_id ON rooms(floor_id);
CREATE INDEX IF NOT EXISTS idx_rooms_door_node_id ON rooms(door_node_id);
CREATE INDEX IF NOT EXISTS idx_connections_from ON connections("from");
CREATE INDEX IF NOT EXISTS idx_connections_to ON connections("to");

-- +goose Down
DROP TABLE IF EXISTS connections CASCADE;
DROP TABLE IF EXISTS rooms CASCADE;
DROP TABLE IF EXISTS nodes CASCADE;
DROP TABLE IF EXISTS floors CASCADE;
DROP TABLE IF EXISTS buildings CASCADE;

