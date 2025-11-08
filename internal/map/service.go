package maps

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetAllBuildings returns all buildings
func (s *Service) GetAllBuildings(ctx context.Context) ([]Building, error) {
	return s.repo.GetAllBuildings(ctx)
}

// GetRoomsByBuilding returns all rooms in a building
func (s *Service) GetRoomsByBuilding(ctx context.Context, buildingID int) ([]Room, error) {
	return s.repo.GetRoomsByBuilding(ctx, buildingID)
}

// FindShortestPath finds the shortest path between two rooms by room names
func (s *Service) FindShortestPath(ctx context.Context, startRoom, endRoom string) ([]string, float64, error) {
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
		return nil, 0, fmt.Errorf("start room %s not found in graph", startRoom)
	}
	if _, ok := graph[endRoom]; !ok {
		return nil, 0, fmt.Errorf("end room %s not found in graph", endRoom)
	}

	// Используем вынесенный алгоритм
	return DijkstraAlgorithm(graph, startRoom, endRoom)
}

func (s *Service) AddRoom(ctx context.Context, room Room, connections []ConnectionDTO) error {
	if err := s.repo.SaveRoom(ctx, &room); err != nil {
		return err
	}

	for _, conn := range connections {
		c := Connection{
			RoomFrom: room.Name,
			RoomTo:   conn.RoomTo,
			Distance: conn.Distance,
		}
		if err := s.repo.SaveConnection(ctx, &c); err != nil {
			return err
		}
	}
	return nil
}
