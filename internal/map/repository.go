package maps

import (
	"context"
	"fmt"
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
	rows, err := r.db.Query(ctx, `SELECT id, name, address FROM buildings`)
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

// SearchRooms позволяет искать комнаты по building_id и/или floor
func (r *Repository) SearchRooms(ctx context.Context, buildingID *int, floor *int) ([]Room, error) {
	query := `SELECT id, building_id, floor, name FROM rooms WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if buildingID != nil {
		query += fmt.Sprintf(" AND building_id=$%d", argIndex)
		args = append(args, *buildingID)
		argIndex++
	}

	if floor != nil {
		query += fmt.Sprintf(" AND floor=$%d", argIndex)
		args = append(args, *floor)
		argIndex++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.BuildingID, &room.Floor, &room.Name); err != nil {
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
		INSERT INTO rooms (building_id, floor, name)
		VALUES ($1, $2, $3)
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

func (r *Repository) DeleteRoom(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	return err
}

func (r *Repository) DeleteConnection(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM connections WHERE id = $1`, id)
	return err
}
