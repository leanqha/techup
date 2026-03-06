package maps

import (
	"context"
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

// SearchRooms returns rooms filtered by building_id and/or floor
func (s *Service) SearchRooms(ctx context.Context, buildingID *int, floor *int) ([]Room, error) {
	var room Room
	if buildingID != nil {
		room.BuildingID = *buildingID
	}
	if floor != nil {
		room.Floor = *floor
	}
	return s.repo.SearchRooms(ctx, &room, buildingID != nil, floor != nil)
}

// FindPath finds the shortest path between two rooms by room names
func (s *Service) FindPath(ctx context.Context, startRoom, endRoom string) ([]string, float64, error) {
	conns, err := s.repo.GetConnections(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Строим граф
	graph := make(map[string][]struct {
		to       string
		distance float64
	})
	for _, c := range conns {
		graph[c.RoomFrom] = append(graph[c.RoomFrom], struct {
			to       string
			distance float64
		}{to: c.RoomTo, distance: c.Distance})
		graph[c.RoomTo] = append(graph[c.RoomTo], struct {
			to       string
			distance float64
		}{to: c.RoomFrom, distance: c.Distance})
	}

	// Проверка существования комнат
	if _, ok := graph[startRoom]; !ok {
		return nil, 0, apperrors.NotFound("start room " + startRoom + " not found in graph")
	}
	if _, ok := graph[endRoom]; !ok {
		return nil, 0, apperrors.NotFound("end room " + endRoom + " not found in graph")
	}

	// Используем A* (с нулевой эвристикой эквивалентно Дейкстре).
	return AStarAlgorithm(graph, startRoom, endRoom, nil)
}

func (s *Service) AddRoom(ctx context.Context, name string, buildingID int, floor int) error {
	room := &Room{
		Name:       name,
		BuildingID: buildingID,
		Floor:      floor,
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
