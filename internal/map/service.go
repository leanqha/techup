package maps

import (
	"context"
	"strconv"
	"techup/internal/apperrors"
)

type RepositoryInterface interface {
	GetAllBuildings(ctx context.Context) ([]Building, error)
	GetBuildingByID(ctx context.Context, id int) (*Building, error)
	AddBuilding(ctx context.Context, building *Building) error
	UpdateBuilding(ctx context.Context, building *Building) error
	DeleteBuilding(ctx context.Context, id int) error

	GetRooms(ctx context.Context) ([]Room, error)
	GetRoomByID(ctx context.Context, id int) (*Room, error)
	SearchRooms(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error)
	AddRoom(ctx context.Context, room *Room) error
	UpdateRoom(ctx context.Context, room *Room) error
	DeleteRoom(ctx context.Context, id int) error

	GetConnections(ctx context.Context) ([]Connection, error)
	GetConnectionByID(ctx context.Context, id int) (*Connection, error)
	AddConnection(ctx context.Context, conn *Connection) error
	UpdateConnection(ctx context.Context, conn *Connection) error
	DeleteConnection(ctx context.Context, id int) error
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// GetAllBuildings returns all buildings
func (s *Service) GetAllBuildings(ctx context.Context) ([]Building, error) {
	return s.repo.GetAllBuildings(ctx)
}

// GetBuildingByID returns a building by ID.
func (s *Service) GetBuildingByID(ctx context.Context, id int) (*Building, error) {
	return s.repo.GetBuildingByID(ctx, id)
}

func (s *Service) AddBuilding(ctx context.Context, building Building) error {
	return s.repo.AddBuilding(ctx, &building)
}

func (s *Service) UpdateBuilding(ctx context.Context, building Building) error {
	return s.repo.UpdateBuilding(ctx, &building)
}

func (s *Service) DeleteBuilding(ctx context.Context, id int) error {
	return s.repo.DeleteBuilding(ctx, id)
}

// SearchRooms returns rooms filtered by building_id and/or floor_id
func (s *Service) SearchRooms(ctx context.Context, buildingID *int, floorID *int) ([]Room, error) {
	var room Room
	if buildingID != nil {
		room.BuildingID = *buildingID
	}
	if floorID != nil {
		room.FloorID = *floorID
	}
	return s.repo.SearchRooms(ctx, &room, buildingID != nil, floorID != nil)
}

// FindPath finds the shortest path between two nodes by ID
func (s *Service) FindPath(ctx context.Context, startNode, endNode string) ([]string, float64, error) {
	conns, err := s.repo.GetConnections(ctx)
	if err != nil {
		return nil, 0, err
	}

	graph := make(map[string][]struct {
		to       string
		distance float64
	})
	for _, c := range conns {
		fromKey := strconv.Itoa(c.FromID)
		toKey := strconv.Itoa(c.ToID)
		graph[fromKey] = append(graph[fromKey], struct {
			to       string
			distance float64
		}{to: toKey, distance: c.Distance})
		graph[toKey] = append(graph[toKey], struct {
			to       string
			distance float64
		}{to: fromKey, distance: c.Distance})
	}

	if _, ok := graph[startNode]; !ok {
		return nil, 0, apperrors.NotFound("start node " + startNode + " not found in graph")
	}
	if _, ok := graph[endNode]; !ok {
		return nil, 0, apperrors.NotFound("end node " + endNode + " not found in graph")
	}

	return AStarAlgorithm(graph, startNode, endNode, nil)
}

func (s *Service) AddRoom(ctx context.Context, name string, title string, buildingID int, floorID int, doorNodeID *int) error {
	room := &Room{
		Name:       name,
		Title:      title,
		BuildingID: buildingID,
		FloorID:    floorID,
		DoorNodeID: doorNodeID,
	}
	return s.repo.AddRoom(ctx, room)
}

func (s *Service) UpdateRoom(ctx context.Context, room Room) error {
	return s.repo.UpdateRoom(ctx, &room)
}

func (s *Service) DeleteRoom(ctx context.Context, id int) error {
	return s.repo.DeleteRoom(ctx, id)
}

func (s *Service) GetRooms(ctx context.Context) ([]Room, error) {
	return s.repo.GetRooms(ctx)
}

func (s *Service) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	return s.repo.GetRoomByID(ctx, id)
}

func (s *Service) GetConnections(ctx context.Context) ([]Connection, error) {
	return s.repo.GetConnections(ctx)
}

func (s *Service) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	return s.repo.GetConnectionByID(ctx, id)
}

func (s *Service) AddConnection(ctx context.Context, connection Connection) error {
	return s.repo.AddConnection(ctx, &connection)
}

func (s *Service) UpdateConnection(ctx context.Context, connection Connection) error {
	return s.repo.UpdateConnection(ctx, &connection)
}

func (s *Service) DeleteConnection(ctx context.Context, id int) error {
	return s.repo.DeleteConnection(ctx, id)
}
