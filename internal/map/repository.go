package maps

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetAllBuildings retrieves all buildings from the database
func (r *Repository) GetAllBuildings(ctx context.Context) ([]Building, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, address, floor_count, description FROM buildings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buildings []Building
	for rows.Next() {
		var b Building
		if err := rows.Scan(&b.ID, &b.Name, &b.Address); err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}

// GetRoomsByBuilding retrieves all rooms of a specific building
func (r *Repository) GetRoomsByBuilding(ctx context.Context, buildingID int) ([]Room, error) {
	rows, err := r.db.Query(ctx, `SELECT id, building_id, number, floor, type FROM rooms WHERE building_id=$1`, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.BuildingID, &room.Name, &room.Floor); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// GetConnections retrieves all connections from the database
func (r *Repository) GetConnections(ctx context.Context) ([]Connection, error) {
	rows, err := r.db.Query(ctx, `SELECT id, room_from, room_to, distance, type FROM connections`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []Connection
	for rows.Next() {
		var conn Connection
		if err := rows.Scan(&conn.ID, &conn.RoomFrom, &conn.RoomTo, &conn.Distance, &conn.Type); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

// AddRoom inserts a new room into the database
func (r *Repository) AddRoom(ctx context.Context, room *Room) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO rooms (building_id, number, floor, type)
		VALUES ($1, $2, $3, $4)
	`, room.BuildingID, room.Name, room.Floor)
	return err
}

// AddConnection inserts a new connection between rooms
func (r *Repository) AddConnection(ctx context.Context, conn *Connection) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO connections (room_from, room_to, distance, type)
		VALUES ($1, $2, $3, $4)
	`, conn.RoomFrom, conn.RoomTo, conn.Distance, conn.Type)
	return err
}

func (r *Repository) SaveRoom(ctx context.Context, room *Room) error {
	_, err := r.db.Exec(ctx, `INSERT INTO rooms (name, building_id, floor) VALUES ($1, $2, $3)`, room.Name, room.BuildingID, room.Floor)
	return err
}

func (r *Repository) SaveConnection(ctx context.Context, conn *Connection) error {
	_, err := r.db.Exec(ctx, `INSERT INTO connections (room_from, room_to, distance) VALUES ($1, $2, $3)`, conn.RoomFrom, conn.RoomTo, conn.Distance)
	return err
}
