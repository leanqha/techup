package maps

type GetShortestPathResponse struct {
	Path []string `json:"path"`
	Dist float64  `json:"dist"`
}

type AddRoomRequest struct {
	Name        string          `json:"name" binding:"required"`
	BuildingID  int             `json:"building_id" binding:"required"`
	Floor       int             `json:"floor" binding:"required"`
	Connections []ConnectionDTO `json:"connections"`
}

type ConnectionDTO struct {
	RoomTo   string  `json:"room_to" binding:"required"`
	Distance float64 `json:"distance" binding:"required"`
}
