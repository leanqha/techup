package maps

type GetShortestPathResponse struct {
	Path []int   `json:"path"`
	Dist float64 `json:"dist"`
}
