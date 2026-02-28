package maps

// GetPathResponse represents a computed path and distance.
type GetPathResponse struct {
	Path []string `json:"path"`
	Dist float64  `json:"dist"`
}

// AddRoomRequest describes payload for creating a room.
type AddRoomRequest struct {
	Name       string `json:"name" binding:"required"`
	BuildingID int    `json:"building_id" binding:"required"`
	Floor      int    `json:"floor" binding:"required"`
}

// AddConnectionRequest describes payload for connecting two rooms.
type AddConnectionRequest struct {
	RoomFrom string  `json:"room_from" binding:"required"`
	RoomTo   string  `json:"room_to" binding:"required"`
	Distance float64 `json:"distance" binding:"required"`
	Type     string  `json:"type" binding:"required"`
}

// SearchRoomsRequest describes payload for filtering rooms by building and floor.
type SearchRoomsRequest struct {
	Floor      int `json:"floor" binding:"required"`
	BuildingID int `json:"building_id" binding:"required"`
}
