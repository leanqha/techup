package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testGraph() map[string][]struct {
	to       string
	distance float64
} {
	graph := make(map[string][]struct {
		to       string
		distance float64
	})

	graph["A"] = append(graph["A"], struct {
		to       string
		distance float64
	}{to: "B", distance: 1})
	graph["B"] = append(graph["B"], struct {
		to       string
		distance float64
	}{to: "A", distance: 1})
	graph["B"] = append(graph["B"], struct {
		to       string
		distance float64
	}{to: "C", distance: 2})
	graph["C"] = append(graph["C"], struct {
		to       string
		distance float64
	}{to: "B", distance: 2})
	graph["A"] = append(graph["A"], struct {
		to       string
		distance float64
	}{to: "C", distance: 10})
	graph["C"] = append(graph["C"], struct {
		to       string
		distance float64
	}{to: "A", distance: 10})

	return graph
}

func TestAStarAlgorithm_FindsShortestPath(t *testing.T) {
	path, dist, err := AStarAlgorithm(testGraph(), "A", "C", nil)

	assert.NoError(t, err)
	assert.Equal(t, []string{"A", "B", "C"}, path)
	assert.Equal(t, 3.0, dist)
}

func TestAStarAlgorithm_NoPath(t *testing.T) {
	graph := map[string][]struct {
		to       string
		distance float64
	}{
		"A": {{to: "B", distance: 1}},
		"B": {{to: "A", distance: 1}},
		"C": {},
	}

	path, dist, err := AStarAlgorithm(graph, "A", "C", nil)

	assert.Error(t, err)
	assert.Nil(t, path)
	assert.Equal(t, 0.0, dist)
}

func TestDijkstraAlgorithm_Compatibility(t *testing.T) {
	path, dist, err := DijkstraAlgorithm(testGraph(), "A", "C")

	assert.NoError(t, err)
	assert.Equal(t, []string{"A", "B", "C"}, path)
	assert.Equal(t, 3.0, dist)
}

func TestReconstructPath(t *testing.T) {
	cameFrom := map[string]string{
		"C": "B",
		"B": "A",
	}

	path := reconstructPath(cameFrom, "C")
	assert.Equal(t, []string{"A", "B", "C"}, path)
}
