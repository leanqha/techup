package maps

// GetPathResponse represents a computed path and distance.
type GetPathResponse struct {
	Path []string `json:"path"`
	Dist float64  `json:"dist"`
}

// AddRoomRequest describes payload for creating a room.
type AddRoomRequest struct {
	Name       string `json:"name" binding:"required"`
	Title      string `json:"title" binding:"required"`
	BuildingID int    `json:"building_id" binding:"required"`
	FloorID    int    `json:"floor_id" binding:"required"`
	DoorNodeID *int   `json:"door_node_id"`
}

// AddConnectionRequest describes payload for connecting two nodes.
type AddConnectionRequest struct {
	FromID   int     `json:"from" binding:"required"`
	ToID     int     `json:"to" binding:"required"`
	Distance float64 `json:"distance" binding:"required"`
}

// SearchRoomsRequest describes payload for filtering rooms by building and floor.
type SearchRoomsRequest struct {
	FloorID    int `json:"floor_id" binding:"required"`
	BuildingID int `json:"building_id" binding:"required"`
}

// AddBuildingRequest describes payload for creating a building.
type AddBuildingRequest struct {
	ID    int    `json:"id" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Title string `json:"title" binding:"required"`
}

// UpdateBuildingRequest describes payload for updating a building.
type UpdateBuildingRequest struct {
	Name  string `json:"name" binding:"required"`
	Title string `json:"title" binding:"required"`
}
