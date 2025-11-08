package maps

import (
	"errors"
	"math"
)

// DijkstraAlgorithm находит кратчайший путь между двумя вершинами графа
func DijkstraAlgorithm(graph map[string][]struct {
	to       string
	distance float64
}, start, end string) ([]string, float64, error) {

	dist := make(map[string]float64)
	prev := make(map[string]string)
	visited := make(map[string]bool)

	// Инициализация расстояний
	for node := range graph {
		dist[node] = math.Inf(1)
		prev[node] = ""
	}
	dist[start] = 0

	for len(visited) < len(graph) {
		minNode := ""
		minDist := math.Inf(1)

		// Находим вершину с минимальным расстоянием
		for node, d := range dist {
			if !visited[node] && d < minDist {
				minNode, minDist = node, d
			}
		}

		if minNode == "" {
			break
		}
		visited[minNode] = true

		// Обновляем соседей
		for _, edge := range graph[minNode] {
			alt := dist[minNode] + edge.distance
			if alt < dist[edge.to] {
				dist[edge.to] = alt
				prev[edge.to] = minNode
			}
		}
	}

	if dist[end] == math.Inf(1) {
		return nil, 0, errors.New("no path found")
	}

	// Восстанавливаем путь
	path := []string{}
	for u := end; u != ""; u = prev[u] {
		path = append([]string{u}, path...)
		if u == start {
			break
		}
	}

	return path, dist[end], nil
}
