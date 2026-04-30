package maps

// Building represents a university building.
type Building struct {
	ID    int    `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Title string `db:"title" json:"title"`
}

// Floor represents a building floor.
type Floor struct {
	ID         int `db:"id" json:"id"`
	BuildingID int `db:"building_id" json:"building_id"`
	Number     int `db:"number" json:"number"`
}

// Node represents a routing point on a floor map.
type Node struct {
	ID         int    `db:"id" json:"id"`
	BuildingID int    `db:"building_id" json:"building_id"`
	FloorID    int    `db:"floor_id" json:"floor_id"`
	X          int    `db:"x" json:"x"`
	Y          int    `db:"y" json:"y"`
	Type       string `db:"type" json:"type"`
}

// Room represents a specific classroom or office.
type Room struct {
	ID         int    `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	Title      string `db:"title" json:"title"`
	BuildingID int    `db:"building_id" json:"building_id"`
	FloorID    int    `db:"floor_id" json:"floor_id"`
	DoorNodeID *int   `db:"door_node_id" json:"door_node_id,omitempty"`
}

// Connection represents a link between two nodes.
type Connection struct {
	ID       int     `db:"id" json:"id"`
	FromID   int     `db:"from" json:"from"`
	ToID     int     `db:"to" json:"to"`
	Distance float64 `db:"distance" json:"distance"`
}
