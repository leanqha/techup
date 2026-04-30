package maps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"techup/internal/apperrors"

	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	GetAllBuildings(ctx context.Context) ([]Building, error)
	GetBuildingByID(ctx context.Context, id int) (*Building, error)
	AddBuilding(ctx context.Context, building Building) error
	UpdateBuilding(ctx context.Context, building Building) error
	DeleteBuilding(ctx context.Context, id int) error

	GetRooms(ctx context.Context) ([]Room, error)
	GetRoomByID(ctx context.Context, id int) (*Room, error)
	FindPath(ctx context.Context, startRoom, endRoom string) ([]string, float64, error)
	SearchRooms(ctx context.Context, buildingID *int, floorID *int) ([]Room, error)
	AddRoom(ctx context.Context, name string, title string, buildingID int, floorID int, doorNodeID *int) error
	UpdateRoom(ctx context.Context, room Room) error
	DeleteRoom(ctx context.Context, id int) error

	GetConnections(ctx context.Context) ([]Connection, error)
	GetConnectionByID(ctx context.Context, id int) (*Connection, error)
	AddConnection(ctx context.Context, connection Connection) error
	UpdateConnection(ctx context.Context, connection Connection) error
	DeleteConnection(ctx context.Context, id int) error
}

// Handler handles HTTP requests for the map module.
type Handler struct {
	service ServiceInterface
}

// NewHandler creates a new Handler with the given Service.
func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// GetBuildings godoc
// @Summary Get all buildings
// @Description Returns a list of all buildings
// @Tags map
// @Produce json
// @Success 200 {array} Building
// @Failure 500 {object} error
// @Router /buildings [get]
func (h *Handler) GetBuildings(c *gin.Context) {
	buildings, err := h.service.GetAllBuildings(c.Request.Context())
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, buildings)
}

// GetBuilding godoc
// @Summary Get a building by ID
// @Description Returns a building by its ID
// @Tags map
// @Produce json
// @Param id path int true "Building ID"
// @Success 200 {object} Building
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /buildings/{id} [get]
func (h *Handler) GetBuilding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	building, err := h.service.GetBuildingByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, building)
}

// AddBuilding godoc
// @Summary Add a new building
// @Description Creates a new building
// @Tags map
// @Accept json
// @Produce json
// @Param building body AddBuildingRequest true "Building data"
// @Success 201 {object} Building
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /buildings [post]
func (h *Handler) AddBuilding(c *gin.Context) {
	var req AddBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddBuilding(c.Request.Context(), Building{
		ID:    req.ID,
		Name:  req.Name,
		Title: req.Title,
	}); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusCreated)
}

// UpdateBuilding godoc
// @Summary Update an existing building
// @Description Updates a building by its ID
// @Tags map
// @Accept json
// @Produce json
// @Param id path int true "Building ID"
// @Param building body UpdateBuildingRequest true "Updated building data"
// @Success 200 {object} Building
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /buildings/{id} [put]
func (h *Handler) UpdateBuilding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateBuildingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateBuilding(c.Request.Context(), Building{
		ID:    id,
		Name:  req.Name,
		Title: req.Title,
	}); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteBuilding godoc
// @Summary Delete a building
// @Description Deletes a building by its ID
// @Tags map
// @Produce json
// @Param id path int true "Building ID"
// @Success 204 {object} nil
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /buildings/{id} [delete]
func (h *Handler) DeleteBuilding(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteBuilding(c.Request.Context(), id); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetRooms godoc
// @Summary Get all rooms
// @Description Returns a list of all rooms
// @Tags map
// @Produce json
// @Success 200 {array} Room
// @Failure 500 {object} error
// @Router /rooms [get]
func (h *Handler) GetRooms(c *gin.Context) {
	rooms, err := h.service.GetRooms(c.Request.Context())
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

// GetRoom godoc
// @Summary Get a room by ID
// @Description Returns a room by its ID
// @Tags map
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {object} Room
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /rooms/{id} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	room, err := h.service.GetRoomByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, room)
}

// GetPath godoc
// @Summary Get path between two rooms
// @Description Calculates path between start_room_id and end_room_id
// @Tags map
// @Produce json
// @Param start_room_id query int true "Start Room ID"
// @Param end_room_id query int true "End Room ID"
// @Success 200 {object} GetPathResponse
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /shortest-path [get]
func (h *Handler) GetPath(c *gin.Context) {
	startRoom := c.Param("start")
	endRoom := c.Param("end")

	path, distance, err := h.service.FindPath(c.Request.Context(), startRoom, endRoom)
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	response := GetPathResponse{
		Path: path,
		Dist: distance,
	}
	c.JSON(http.StatusOK, response)
}

// SearchRooms godoc
// @Summary Search rooms
// @Description Returns rooms filtered by building_id and/or floor_id
// @Tags map
// @Produce json
// @Param building_id query int false "Building ID"
// @Param floor_id query int false "Floor ID"
// @Success 200 {array} Room
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/v1/map/search [get]
func (h *Handler) SearchRooms(c *gin.Context) {
	var (
		buildingID *int
		floorID    *int
	)

	if b := c.Query("building_id"); b != "" {
		var id int
		if _, err := fmt.Sscanf(b, "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid building_id"})
			return
		}
		buildingID = &id
	}

	if f := c.Query("floor_id"); f != "" {
		var fl int
		if _, err := fmt.Sscanf(f, "%d", &fl); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid floor_id"})
			return
		}
		floorID = &fl
	}

	rooms, err := h.service.SearchRooms(c.Request.Context(), buildingID, floorID)
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) AddRoom(c *gin.Context) {
	var req AddRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddRoom(c.Request.Context(), req.Name, req.Title, req.BuildingID, req.FloorID, req.DoorNodeID); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) UpdateRoom(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req AddRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateRoom(c.Request.Context(), Room{
		ID:         id,
		Name:       req.Name,
		Title:      req.Title,
		BuildingID: req.BuildingID,
		FloorID:    req.FloorID,
		DoorNodeID: req.DoorNodeID,
	}); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) DeleteRoom(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteRoom(c.Request.Context(), id); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetConnections godoc
// @Summary Get all connections
// @Description Returns a list of all connections
// @Tags map
// @Produce json
// @Success 200 {array} Connection
// @Failure 500 {object} error
// @Router /connections [get]
func (h *Handler) GetConnections(c *gin.Context) {
	connections, err := h.service.GetConnections(c.Request.Context())
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, connections)
}

// GetConnection godoc
// @Summary Get a connection by ID
// @Description Returns a connection by its ID
// @Tags map
// @Produce json
// @Param id path int true "Connection ID"
// @Success 200 {object} Connection
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /connections/{id} [get]
func (h *Handler) GetConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	connection, err := h.service.GetConnectionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}
	c.JSON(http.StatusOK, connection)
}

// AddConnection godoc
// @Summary Add a new connection
// @Description Creates a new connection
// @Tags map
// @Accept json
// @Produce json
// @Param connection body AddConnectionRequest true "Connection data"
// @Success 201
// @Failure 400 {object} error
// @Failure 500 {object} error
// @Router /admin/connection [post]
func (h *Handler) AddConnection(c *gin.Context) {
	var req AddConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddConnection(c.Request.Context(), Connection{
		FromID:   req.FromID,
		ToID:     req.ToID,
		Distance: req.Distance,
	}); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusCreated)
}

// UpdateConnection godoc
// @Summary Update an existing connection
// @Description Updates a connection by its ID
// @Tags map
// @Accept json
// @Produce json
// @Param id path int true "Connection ID"
// @Param connection body AddConnectionRequest true "Updated connection data"
// @Success 200 "OK"
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /api/v1/admin/connection/{id} [put]
func (h *Handler) UpdateConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req AddConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateConnection(c.Request.Context(), Connection{
		ID:       id,
		FromID:   req.FromID,
		ToID:     req.ToID,
		Distance: req.Distance,
	}); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteConnection godoc
// @Summary Delete a connection
// @Description Deletes a connection by its ID
// @Tags map
// @Produce json
// @Param id path int true "Connection ID"
// @Success 204 "No Content"
// @Failure 400 {object} error
// @Failure 404 {object} error
// @Failure 500 {object} error
// @Router /api/v1/admin/connection/{id} [delete]
func (h *Handler) DeleteConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteConnection(c.Request.Context(), id); err != nil {
		c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
		return
	}

	c.Status(http.StatusNoContent)
}
