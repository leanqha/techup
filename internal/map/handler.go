package maps

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
	startRoom := c.Param("start")
	endRoom := c.Param("end")

	path, distance, err := h.Service.FindShortestPath(c.Request.Context(), startRoom, endRoom)
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

func (h *Handler) AddRoom(c *gin.Context) {
	var req AddRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := Room{
		Name:       req.Name,
		BuildingID: req.BuildingID,
		Floor:      req.Floor,
	}

	if err := h.Service.AddRoom(c.Request.Context(), room, req.Connections); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "room added"})
}
