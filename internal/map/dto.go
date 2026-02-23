package maps

type GetPathResponse struct {
	Path []string `json:"path"`
	Dist float64  `json:"dist"`
}

type AddRoomRequest struct {
	Name       string `json:"name" binding:"required"`
	BuildingID int    `json:"building_id" binding:"required"`
	Floor      int    `json:"floor" binding:"required"`
}

type AddConnectionRequest struct {
	RoomFrom string  `json:"room_from" binding:"required"`
	RoomTo   string  `json:"room_to" binding:"required"`
	Distance float64 `json:"distance" binding:"required"`
	Type     string  `json:"type" binding:"required"`
}

type SearchRoomsRequest struct {
	Floor      int `json:"floor" binding:"required"`
	BuildingID int `json:"building_id" binding:"required"`
}
