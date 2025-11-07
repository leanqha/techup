package maps

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the map module.
type Handler struct {
	Service *Service
}

// NewHandler creates a new Handler with the given Service.
func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// GetBuildings godoc
// @Summary Get all buildings
// @Description Returns a list of all buildings
// @Tags map
// @Produce json
// @Success 200 {array} Building
// @Failure 500 {object} gin.H{"error": string}
// @Router /buildings [get]
func (h *Handler) GetBuildings(c *gin.Context) {
	buildings, err := h.Service.GetAllBuildings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get buildings"})
		return
	}
	c.JSON(http.StatusOK, buildings)
}

// GetRoomsByBuilding godoc
// @Summary Get rooms by building ID
// @Description Returns a list of rooms for a given building ID
// @Tags map
// @Produce json
// @Param building_id path int true "Building ID"
// @Success 200 {array} Room
// @Failure 400 {object} gin.H{"error": string}
// @Failure 404 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /buildings/{building_id}/rooms [get]
func (h *Handler) GetRoomsByBuilding(c *gin.Context) {
	buildingIDStr := c.Param("building_id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	rooms, err := h.Service.GetRoomsByBuilding(c.Request.Context(), buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rooms"})
		return
	}
	if rooms == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Building not found"})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

// GetShortestPath godoc
// @Summary Get shortest path between two rooms
// @Description Calculates the shortest path between start_room_id and end_room_id
// @Tags map
// @Produce json
// @Param start_room_id query int true "Start Room ID"
// @Param end_room_id query int true "End Room ID"
// @Success 200 {object} GetShortestPathResponse
// @Failure 400 {object} gin.H{"error": string}
// @Failure 500 {object} gin.H{"error": string}
// @Router /shortest-path [get]
func (h *Handler) GetShortestPath(c *gin.Context) {
	startIDStr := c.Query("start_room_id")
	endIDStr := c.Query("end_room_id")

	startID, err := strconv.Atoi(startIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_room_id"})
		return
	}
	endID, err := strconv.Atoi(endIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_room_id"})
		return
	}

	path, distance, err := h.Service.FindShortestPath(c.Request.Context(), startID, endID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate shortest path"})
		return
	}
	response := GetShortestPathResponse{
		Path: path,
		Dist: distance,
	}
	c.JSON(http.StatusOK, response)
}
