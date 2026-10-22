package maps

import (
	"context"
	"strconv"
	"techup/config"
	"techup/internal/apperrors"
	"techup/internal/cache"
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

	// Reference data read on nearly every map request but written only through
	// admin CRUD. Each cache is invalidated by the matching write below; the
	// navigation graph is cached pre-built so pathfinding skips both the DB
	// read and the graph reconstruction. Nil when caching is disabled.
	buildingsCache *cache.Value[[]Building]
	roomsCache     *cache.Value[[]Room]
	connsCache     *cache.Value[[]Connection]
	graphCache     *cache.Value[adjacency]
}

func NewService(repo RepositoryInterface) *Service {
	s := &Service{repo: repo}
	if ttl := config.GetReferenceCacheTTL(); ttl > 0 {
		s.buildingsCache = cache.New[[]Building](ttl)
		s.roomsCache = cache.New[[]Room](ttl)
		s.connsCache = cache.New[[]Connection](ttl)
		s.graphCache = cache.New[adjacency](ttl)
	}
	return s
}

// GetAllBuildings returns all buildings
func (s *Service) GetAllBuildings(ctx context.Context) ([]Building, error) {
	return s.buildingsCache.Get(func() ([]Building, error) {
		return s.repo.GetAllBuildings(ctx)
	})
}

// GetBuildingByID returns a building by ID.
func (s *Service) GetBuildingByID(ctx context.Context, id int) (*Building, error) {
	return s.repo.GetBuildingByID(ctx, id)
}

func (s *Service) AddBuilding(ctx context.Context, building Building) error {
	if err := s.repo.AddBuilding(ctx, &building); err != nil {
		return err
	}
	s.buildingsCache.Invalidate()
	return nil
}

func (s *Service) UpdateBuilding(ctx context.Context, building Building) error {
	if err := s.repo.UpdateBuilding(ctx, &building); err != nil {
		return err
	}
	s.buildingsCache.Invalidate()
	return nil
}

func (s *Service) DeleteBuilding(ctx context.Context, id int) error {
	if err := s.repo.DeleteBuilding(ctx, id); err != nil {
		return err
	}
	s.buildingsCache.Invalidate()
	return nil
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

// FindPath finds the shortest path between two nodes by ID. The adjacency graph
// is cached pre-built, so a warm request skips both the connections query and
// the graph reconstruction and goes straight to A*.
func (s *Service) FindPath(ctx context.Context, startNode, endNode string) ([]string, float64, error) {
	graph, err := s.buildGraph(ctx)
	if err != nil {
		return nil, 0, err
	}

	if _, ok := graph[startNode]; !ok {
		return nil, 0, apperrors.NotFound("start node " + startNode + " not found in graph")
	}
	if _, ok := graph[endNode]; !ok {
		return nil, 0, apperrors.NotFound("end node " + endNode + " not found in graph")
	}

	return AStarAlgorithm(graph, startNode, endNode, nil)
}

// buildGraph returns the undirected adjacency graph of all connections, built
// once and cached until a connection changes.
func (s *Service) buildGraph(ctx context.Context) (adjacency, error) {
	return s.graphCache.Get(func() (adjacency, error) {
		conns, err := s.GetConnections(ctx)
		if err != nil {
			return nil, err
		}

		graph := make(adjacency, len(conns))
		for _, c := range conns {
			fromKey := strconv.Itoa(c.FromID)
			toKey := strconv.Itoa(c.ToID)
			graph[fromKey] = append(graph[fromKey], edge{to: toKey, distance: c.Distance})
			graph[toKey] = append(graph[toKey], edge{to: fromKey, distance: c.Distance})
		}
		return graph, nil
	})
}

func (s *Service) AddRoom(ctx context.Context, name string, title string, buildingID int, floorID int, doorNodeID *int) error {
	room := &Room{
		Name:       name,
		Title:      title,
		BuildingID: buildingID,
		FloorID:    floorID,
		DoorNodeID: doorNodeID,
	}
	if err := s.repo.AddRoom(ctx, room); err != nil {
		return err
	}
	s.roomsCache.Invalidate()
	return nil
}

func (s *Service) UpdateRoom(ctx context.Context, room Room) error {
	if err := s.repo.UpdateRoom(ctx, &room); err != nil {
		return err
	}
	s.roomsCache.Invalidate()
	return nil
}

func (s *Service) DeleteRoom(ctx context.Context, id int) error {
	if err := s.repo.DeleteRoom(ctx, id); err != nil {
		return err
	}
	s.roomsCache.Invalidate()
	return nil
}

func (s *Service) GetRooms(ctx context.Context) ([]Room, error) {
	return s.roomsCache.Get(func() ([]Room, error) {
		return s.repo.GetRooms(ctx)
	})
}

func (s *Service) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	return s.repo.GetRoomByID(ctx, id)
}

func (s *Service) GetConnections(ctx context.Context) ([]Connection, error) {
	return s.connsCache.Get(func() ([]Connection, error) {
		return s.repo.GetConnections(ctx)
	})
}

func (s *Service) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	return s.repo.GetConnectionByID(ctx, id)
}

// invalidateConnections drops both the raw connection list and the derived
// navigation graph; any connection write changes both.
func (s *Service) invalidateConnections() {
	s.connsCache.Invalidate()
	s.graphCache.Invalidate()
}

func (s *Service) AddConnection(ctx context.Context, connection Connection) error {
	if err := s.repo.AddConnection(ctx, &connection); err != nil {
		return err
	}
	s.invalidateConnections()
	return nil
}

func (s *Service) UpdateConnection(ctx context.Context, connection Connection) error {
	if err := s.repo.UpdateConnection(ctx, &connection); err != nil {
		return err
	}
	s.invalidateConnections()
	return nil
}

func (s *Service) DeleteConnection(ctx context.Context, id int) error {
	if err := s.repo.DeleteConnection(ctx, id); err != nil {
		return err
	}
	s.invalidateConnections()
	return nil
}
