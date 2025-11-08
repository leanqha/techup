package maps

import (
	"context"
	"errors"
	"fmt"
	"math"
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

	// Build graph using room names
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

	// Validate start and end rooms exist
	if _, ok := graph[startRoom]; !ok {
		return nil, 0, fmt.Errorf("start room %s not found in graph", startRoom)
	}
	if _, ok := graph[endRoom]; !ok {
		return nil, 0, fmt.Errorf("end room %s not found in graph", endRoom)
	}

	dist := make(map[string]float64)
	prev := make(map[string]string)
	visited := make(map[string]bool)

	for node := range graph {
		dist[node] = math.Inf(1)
		prev[node] = ""
	}
	dist[startRoom] = 0

	// Dijkstra's algorithm
	for len(visited) < len(graph) {
		minNode := ""
		minDist := math.Inf(1)
		for node, d := range dist {
			if !visited[node] && d < minDist {
				minNode, minDist = node, d
			}
		}
		if minNode == "" {
			break
		}

		visited[minNode] = true

		for _, edge := range graph[minNode] {
			alt := dist[minNode] + edge.distance
			if alt < dist[edge.to] {
				dist[edge.to] = alt
				prev[edge.to] = minNode
			}
		}
	}

	if dist[endRoom] == math.Inf(1) {
		return nil, 0, errors.New("no path found")
	}

	// Reconstruct path
	path := []string{}
	for u := endRoom; u != ""; u = prev[u] {
		path = append([]string{u}, path...)
		if u == startRoom {
			break
		}
	}

	return path, dist[endRoom], nil
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
