package maps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	GetAllBuildings(ctx context.Context) ([]Building, error)
	FindPath(ctx context.Context, startRoom, endRoom string) ([]string, float64, error)
	SearchRooms(ctx context.Context, buildingID *int, floor *int) ([]Room, error)
	AddRoom(ctx context.Context, name string, buildingID int, floor int) error
	UpdateRoom(ctx context.Context, room Room) error
	DeleteRoom(ctx context.Context, id int) error
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get buildings"})
		return
	}
	c.JSON(http.StatusOK, buildings)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate shortest path"})
		return
	}
	response := GetPathResponse{
		Path: path,
		Dist: distance,
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) AddRoom(c *gin.Context) {
	var req AddRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddRoom(c.Request.Context(), req.Name, req.BuildingID, req.Floor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		BuildingID: req.BuildingID,
		Floor:      req.Floor,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchRooms godoc
// @Summary Search rooms
// @Description Returns rooms filtered by building_id and/or floor
// @Tags map
// @Produce json
// @Param building_id query int false "Building ID"
// @Param floor query int false "Floor number"
// @Success 200 {array} Room
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/v1/map/search [get]
func (h *Handler) SearchRooms(c *gin.Context) {
	var (
		buildingID *int
		floor      *int
	)

	if b := c.Query("building_id"); b != "" {
		var id int
		if _, err := fmt.Sscanf(b, "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid building_id"})
			return
		}
		buildingID = &id
	}

	if f := c.Query("floor"); f != "" {
		var fl int
		if _, err := fmt.Sscanf(f, "%d", &fl); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid floor"})
			return
		}
		floor = &fl
	}

	rooms, err := h.service.SearchRooms(c.Request.Context(), buildingID, floor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) AddConnection(c *gin.Context) {
	var req AddConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddConnection(c.Request.Context(), Connection{
		RoomFrom: req.RoomFrom,
		RoomTo:   req.RoomTo,
		Distance: req.Distance,
		Type:     req.Type,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

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
		RoomFrom: req.RoomFrom,
		RoomTo:   req.RoomTo,
		Distance: req.Distance,
		Type:     req.Type,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) DeleteConnection(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteConnection(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
