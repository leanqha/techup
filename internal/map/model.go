package maps

// Building represents a university building
type Building struct {
	ID      int    `db:"id" json:"id"`
	Name    string `db:"name" json:"name"`
	Address string `db:"address" json:"address"`
}

// Room represents a specific classroom, office, or corridor
type Room struct {
	ID         int    `db:"id" json:"id"`
	BuildingID int    `db:"building_id" json:"building_id"`
	Floor      int    `db:"floor" json:"floor"`
	Name       string `db:"name" json:"name"`
}

// Connection represents a link between two rooms (corridor, stairs, etc.)
type Connection struct {
	ID       int     `db:"id" json:"id"`
	RoomFrom string  `db:"room_from" json:"room_from"`
	RoomTo   string  `db:"room_to" json:"room_to"`
	Distance float64 `db:"distance" json:"distance"`
	Type     string  `db:"type" json:"type"` // corridor / stairs
}
