package maps

import (
	"context"
	"database/sql"
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
	query := `SELECT id, name, title FROM buildings`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetAllBuildings")
		return nil, err
	}
	defer rows.Close()

	var buildings []Building
	for rows.Next() {
		var b Building
		if err := rows.Scan(&b.ID, &b.Name, &b.Title); err != nil {
			return nil, err
		}
		buildings = append(buildings, b)
	}
	return buildings, nil
}

// SearchRooms позволяет искать комнаты по building_id и/или floor
func (r *Repository) SearchRooms(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error) {
	query := `SELECT id, name, title, building_id, floor_id, door_node_id FROM rooms WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if hasBuildingID {
		query += fmt.Sprintf(" AND building_id=$%d", argIndex)
		args = append(args, room.BuildingID)
		argIndex++
	}

	if hasFloor {
		query += fmt.Sprintf(" AND floor_id=$%d", argIndex)
		args = append(args, room.FloorID)
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
		var doorNodeID sql.NullInt64
		if err := rows.Scan(&room.ID, &room.Name, &room.Title, &room.BuildingID, &room.FloorID, &doorNodeID); err != nil {
			return nil, err
		}
		room.DoorNodeID = nullIntToPtr(doorNodeID)
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetConnections retrieves all connections from the database
func (r *Repository) GetConnections(ctx context.Context) ([]Connection, error) {
	query := `SELECT id, "from", "to", distance FROM connections`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetConnections")
		return nil, err
	}
	defer rows.Close()

	var connections []Connection
	for rows.Next() {
		var conn Connection
		if err := rows.Scan(&conn.ID, &conn.FromID, &conn.ToID, &conn.Distance); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}
	return connections, nil
}

// AddRoom inserts a new room into the database
func (r *Repository) AddRoom(ctx context.Context, room *Room) error {
	query := `
		INSERT INTO rooms (name, title, building_id, floor_id, door_node_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query, room.Name, room.Title, room.BuildingID, room.FloorID, ptrToNullInt(room.DoorNodeID))
	if err != nil {
		logger.LogSQLError(err, query, room.Name, room.Title, room.BuildingID, room.FloorID, room.DoorNodeID)
	}
	return err
}

func (r *Repository) UpdateRoom(ctx context.Context, room *Room) error {
	query := `
		UPDATE rooms
		SET name = $1, title = $2, building_id = $3, floor_id = $4, door_node_id = $5
		WHERE id = $6
	`
	ct, err := r.db.Exec(ctx, query, room.Name, room.Title, room.BuildingID, room.FloorID, ptrToNullInt(room.DoorNodeID), room.ID)
	if err != nil {
		logger.LogSQLError(err, query, room.Name, room.Title, room.BuildingID, room.FloorID, room.DoorNodeID, room.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("room not found")
	}
	return nil
}

// AddConnection inserts a new connection between nodes
func (r *Repository) AddConnection(ctx context.Context, conn *Connection) error {
	query := `
		INSERT INTO connections ("from", "to", distance)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, conn.FromID, conn.ToID, conn.Distance)
	if err != nil {
		logger.LogSQLError(err, query, conn.FromID, conn.ToID, conn.Distance)
	}
	return err
}

func (r *Repository) UpdateConnection(ctx context.Context, conn *Connection) error {
	query := `
		UPDATE connections
		SET "from" = $1, "to" = $2, distance = $3, updated_at = NOW()
		WHERE id = $4
	`
	ct, err := r.db.Exec(ctx, query, conn.FromID, conn.ToID, conn.Distance, conn.ID)
	if err != nil {
		logger.LogSQLError(err, query, conn.FromID, conn.ToID, conn.Distance, conn.ID)
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
	query := `SELECT id, name, title FROM buildings WHERE id = $1`
	var b Building
	if err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.Name, &b.Title); err != nil {
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
		INSERT INTO buildings (id, name, title)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, building.ID, building.Name, building.Title)
	if err != nil {
		logger.LogSQLError(err, query, building.ID, building.Name, building.Title)
	}
	return err
}

func (r *Repository) UpdateBuilding(ctx context.Context, building *Building) error {
	query := `
		UPDATE buildings
		SET name = $1, title = $2, updated_at = NOW()
		WHERE id = $3
	`
	ct, err := r.db.Exec(ctx, query, building.Name, building.Title, building.ID)
	if err != nil {
		logger.LogSQLError(err, query, building.Name, building.Title, building.ID)
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
	query := `SELECT id, name, title, building_id, floor_id, door_node_id FROM rooms`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "GetRooms")
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		var doorNodeID sql.NullInt64
		if err := rows.Scan(&room.ID, &room.Name, &room.Title, &room.BuildingID, &room.FloorID, &doorNodeID); err != nil {
			return nil, err
		}
		room.DoorNodeID = nullIntToPtr(doorNodeID)
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *Repository) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	query := `SELECT id, name, title, building_id, floor_id, door_node_id FROM rooms WHERE id = $1`
	var room Room
	var doorNodeID sql.NullInt64
	if err := r.db.QueryRow(ctx, query, id).Scan(&room.ID, &room.Name, &room.Title, &room.BuildingID, &room.FloorID, &doorNodeID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("room not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	room.DoorNodeID = nullIntToPtr(doorNodeID)
	return &room, nil
}

func (r *Repository) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	query := `SELECT id, "from", "to", distance FROM connections WHERE id = $1`
	var conn Connection
	if err := r.db.QueryRow(ctx, query, id).Scan(&conn.ID, &conn.FromID, &conn.ToID, &conn.Distance); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("connection not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &conn, nil
}

func nullIntToPtr(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}

func ptrToNullInt(value *int) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*value), Valid: true}
}
