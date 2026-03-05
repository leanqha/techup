package maps

import (
	"context"
	"fmt"
	"techup/internal/apperrors"
	"techup/internal/logger"

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
	query := `SELECT id, name, address FROM buildings`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetAllBuildings")
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
func (r *Repository) SearchRooms(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error) {
	query := `SELECT id, building_id, floor, name FROM rooms WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if hasBuildingID {
		query += fmt.Sprintf(" AND building_id=$%d", argIndex)
		args = append(args, room.BuildingID)
		argIndex++
	}

	if hasFloor {
		query += fmt.Sprintf(" AND floor=$%d", argIndex)
		args = append(args, room.Floor)
		argIndex++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.LogSQLError(err, query, args...)
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
	query := `SELECT id, room_from, room_to, distance, type FROM connections`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetConnections")
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
	query := `
		INSERT INTO rooms (building_id, floor, name)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, room.BuildingID, room.Floor, room.Name)
	if err != nil {
		logger.LogSQLError(err, query, room.BuildingID, room.Floor, room.Name)
	}
	return err
}

func (r *Repository) UpdateRoom(ctx context.Context, room *Room) error {
	query := `
		UPDATE rooms
		SET building_id = $1, floor = $2, name = $3, updated_at = NOW()
		WHERE id = $4
	`
	ct, err := r.db.Exec(ctx, query, room.BuildingID, room.Floor, room.Name, room.ID)
	if err != nil {
		logger.LogSQLError(err, query, room.BuildingID, room.Floor, room.Name, room.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("room not found")
	}
	return nil
}

// AddConnection inserts a new connection between rooms
func (r *Repository) AddConnection(ctx context.Context, conn *Connection) error {
	query := `
		INSERT INTO connections (room_from, room_to, distance, type)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, conn.RoomFrom, conn.RoomTo, conn.Distance, conn.Type)
	if err != nil {
		logger.LogSQLError(err, query, conn.RoomFrom, conn.RoomTo, conn.Distance, conn.Type)
	}
	return err
}

func (r *Repository) UpdateConnection(ctx context.Context, conn *Connection) error {
	query := `
		UPDATE connections
		SET room_from = $1, room_to = $2, distance = $3, type = $4, updated_at = NOW()
		WHERE id = $5
	`
	ct, err := r.db.Exec(ctx, query, conn.RoomFrom, conn.RoomTo, conn.Distance, conn.Type, conn.ID)
	if err != nil {
		logger.LogSQLError(err, query, conn.RoomFrom, conn.RoomTo, conn.Distance, conn.Type, conn.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("connection not found")
	}
	return nil
}

func (r *Repository) DeleteRoom(ctx context.Context, id int) error {
	query := `DELETE FROM rooms WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("room not found")
	}
	return nil
}

// DeleteConnection deletes connection by ID
func (r *Repository) DeleteConnection(ctx context.Context, id int) error {
	query := `DELETE FROM connections WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("connection not found")
	}
	return nil
}
