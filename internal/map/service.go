package maps

import (
	"context"
	"errors"
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

// FindShortestPath finds the shortest path between two rooms using Dijkstra's algorithm
func (s *Service) FindShortestPath(ctx context.Context, startRoom, endRoom int) ([]int, float64, error) {
	conns, err := s.repo.GetConnections(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Build graph
	graph := make(map[int][]struct {
		to       int
		distance float64
	})
	for _, c := range conns {
		graph[c.RoomFrom] = append(graph[c.RoomFrom], struct {
			to       int
			distance float64
		}{to: c.RoomTo, distance: c.Distance})
		graph[c.RoomTo] = append(graph[c.RoomTo], struct {
			to       int
			distance float64
		}{to: c.RoomFrom, distance: c.Distance})
	}

	dist := make(map[int]float64)
	prev := make(map[int]int)
	visited := make(map[int]bool)

	for node := range graph {
		dist[node] = math.Inf(1)
	}
	dist[startRoom] = 0

	for {
		// find node with smallest distance
		minNode := -1
		minDist := math.Inf(1)
		for node, d := range dist {
			if !visited[node] && d < minDist {
				minNode, minDist = node, d
			}
		}
		if minNode == -1 {
			break
		}
		if minNode == endRoom {
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

	if _, ok := dist[endRoom]; !ok || dist[endRoom] == math.Inf(1) {
		return nil, 0, errors.New("no path found")
	}

	// reconstruct path
	path := []int{}
	for u := endRoom; u != 0; u = prev[u] {
		path = append([]int{u}, path...)
		if u == startRoom {
			break
		}
	}

	return path, dist[endRoom], nil
}
