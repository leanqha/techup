package maps

import (
	"context"
	"fmt"
	"techup/internal/apperrors"
	"techup/internal/logger"

	"github.com/jackc/pgx/v5"
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

func (r *Repository) GetBuildingByID(ctx context.Context, id int) (*Building, error) {
	query := `SELECT id, name, address FROM buildings WHERE id = $1`
	var b Building
	if err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.Name, &b.Address); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("building not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &b, nil
}

func (r *Repository) AddBuilding(ctx context.Context, building *Building) error {
	query := `
		INSERT INTO buildings (id, name, address)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, building.ID, building.Name, building.Address)
	if err != nil {
		logger.LogSQLError(err, query, building.ID, building.Name, building.Address)
	}
	return err
}

func (r *Repository) UpdateBuilding(ctx context.Context, building *Building) error {
	query := `
		UPDATE buildings
		SET name = $1, address = $2, updated_at = NOW()
		WHERE id = $3
	`
	ct, err := r.db.Exec(ctx, query, building.Name, building.Address, building.ID)
	if err != nil {
		logger.LogSQLError(err, query, building.Name, building.Address, building.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("building not found")
	}
	return nil
}

func (r *Repository) DeleteBuilding(ctx context.Context, id int) error {
	query := `DELETE FROM buildings WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("building not found")
	}
	return nil
}

func (r *Repository) GetRooms(ctx context.Context) ([]Room, error) {
	query := `SELECT id, building_id, floor, name FROM rooms`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetRooms")
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

func (r *Repository) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	query := `SELECT id, building_id, floor, name FROM rooms WHERE id = $1`
	var room Room
	if err := r.db.QueryRow(ctx, query, id).Scan(&room.ID, &room.BuildingID, &room.Floor, &room.Name); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("room not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &room, nil
}

func (r *Repository) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	query := `SELECT id, room_from, room_to, distance, type FROM connections WHERE id = $1`
	var conn Connection
	if err := r.db.QueryRow(ctx, query, id).Scan(&conn.ID, &conn.RoomFrom, &conn.RoomTo, &conn.Distance, &conn.Type); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("connection not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &conn, nil
}
