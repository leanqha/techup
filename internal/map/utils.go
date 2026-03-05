package maps

import (
	"container/heap"
	"errors"
)

type priorityNode struct {
	node     string
	priority float64
	index    int
}

type priorityQueue []*priorityNode

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*priorityNode)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// AStarAlgorithm находит кратчайший путь между двумя вершинами графа.
// При нулевой эвристике работает как алгоритм Дейкстры.
func AStarAlgorithm(graph map[string][]struct {
	to       string
	distance float64
}, start, end string, heuristic func(from, to string) float64) ([]string, float64, error) {
	if heuristic == nil {
		heuristic = func(_, _ string) float64 { return 0 }
	}

	gScore := make(map[string]float64, len(graph))
	cameFrom := make(map[string]string, len(graph))

	openSet := make(priorityQueue, 0, len(graph))
	heap.Init(&openSet)
	heap.Push(&openSet, &priorityNode{node: start, priority: heuristic(start, end)})
	gScore[start] = 0

	for openSet.Len() > 0 {
		current := heap.Pop(&openSet).(*priorityNode).node
		if current == end {
			path := reconstructPath(cameFrom, current)
			return path, gScore[end], nil
		}

		for _, edge := range graph[current] {
			tentativeG := gScore[current] + edge.distance
			bestG, seen := gScore[edge.to]
			if !seen || tentativeG < bestG {
				cameFrom[edge.to] = current
				gScore[edge.to] = tentativeG
				fScore := tentativeG + heuristic(edge.to, end)
				heap.Push(&openSet, &priorityNode{node: edge.to, priority: fScore})
			}
		}
	}

	return nil, 0, errors.New("no path found")
}

func reconstructPath(cameFrom map[string]string, current string) []string {
	path := []string{current}
	for {
		prev, ok := cameFrom[current]
		if !ok {
			break
		}
		current = prev
		path = append([]string{current}, path...)
	}
	return path
}

// DijkstraAlgorithm сохранен для совместимости и делегирует вызов A* с нулевой эвристикой.
func DijkstraAlgorithm(graph map[string][]struct {
	to       string
	distance float64
}, start, end string) ([]string, float64, error) {
	return AStarAlgorithm(graph, start, end, nil)
}
