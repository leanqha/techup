package maps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockMapService struct {
	getAllBuildingsFn   func(ctx context.Context) ([]Building, error)
	getBuildingByIDFn   func(ctx context.Context, id int) (*Building, error)
	addBuildingFn       func(ctx context.Context, building Building) error
	updateBuildingFn    func(ctx context.Context, building Building) error
	deleteBuildingFn    func(ctx context.Context, id int) error
	getRoomsFn          func(ctx context.Context) ([]Room, error)
	getRoomByIDFn       func(ctx context.Context, id int) (*Room, error)
	findPathFn          func(ctx context.Context, startRoom, endRoom string) ([]string, float64, error)
	searchRoomsFn       func(ctx context.Context, buildingID *int, floor *int) ([]Room, error)
	addRoomFn           func(ctx context.Context, name string, buildingID int, floor int) error
	updateRoomFn        func(ctx context.Context, room Room) error
	deleteRoomFn        func(ctx context.Context, id int) error
	getConnectionsFn    func(ctx context.Context) ([]Connection, error)
	getConnectionByIDFn func(ctx context.Context, id int) (*Connection, error)
	addConnectionFn     func(ctx context.Context, connection Connection) error
	updateConnectionFn  func(ctx context.Context, connection Connection) error
	deleteConnectionFn  func(ctx context.Context, id int) error
}

func (m *mockMapService) GetAllBuildings(ctx context.Context) ([]Building, error) {
	if m.getAllBuildingsFn != nil {
		return m.getAllBuildingsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapService) GetBuildingByID(ctx context.Context, id int) (*Building, error) {
	if m.getBuildingByIDFn != nil {
		return m.getBuildingByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockMapService) AddBuilding(ctx context.Context, building Building) error {
	if m.addBuildingFn != nil {
		return m.addBuildingFn(ctx, building)
	}
	return nil
}

func (m *mockMapService) UpdateBuilding(ctx context.Context, building Building) error {
	if m.updateBuildingFn != nil {
		return m.updateBuildingFn(ctx, building)
	}
	return nil
}

func (m *mockMapService) DeleteBuilding(ctx context.Context, id int) error {
	if m.deleteBuildingFn != nil {
		return m.deleteBuildingFn(ctx, id)
	}
	return nil
}

func (m *mockMapService) GetRooms(ctx context.Context) ([]Room, error) {
	if m.getRoomsFn != nil {
		return m.getRoomsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapService) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	if m.getRoomByIDFn != nil {
		return m.getRoomByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockMapService) FindPath(ctx context.Context, startRoom, endRoom string) ([]string, float64, error) {
	if m.findPathFn != nil {
		return m.findPathFn(ctx, startRoom, endRoom)
	}
	return nil, 0, nil
}

func (m *mockMapService) SearchRooms(ctx context.Context, buildingID *int, floor *int) ([]Room, error) {
	if m.searchRoomsFn != nil {
		return m.searchRoomsFn(ctx, buildingID, floor)
	}
	return nil, nil
}

func (m *mockMapService) AddRoom(ctx context.Context, name string, buildingID int, floor int) error {
	if m.addRoomFn != nil {
		return m.addRoomFn(ctx, name, buildingID, floor)
	}
	return nil
}

func (m *mockMapService) UpdateRoom(ctx context.Context, room Room) error {
	if m.updateRoomFn != nil {
		return m.updateRoomFn(ctx, room)
	}
	return nil
}

func (m *mockMapService) DeleteRoom(ctx context.Context, id int) error {
	if m.deleteRoomFn != nil {
		return m.deleteRoomFn(ctx, id)
	}
	return nil
}

func (m *mockMapService) GetConnections(ctx context.Context) ([]Connection, error) {
	if m.getConnectionsFn != nil {
		return m.getConnectionsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapService) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	if m.getConnectionByIDFn != nil {
		return m.getConnectionByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockMapService) AddConnection(ctx context.Context, connection Connection) error {
	if m.addConnectionFn != nil {
		return m.addConnectionFn(ctx, connection)
	}
	return nil
}

func (m *mockMapService) UpdateConnection(ctx context.Context, connection Connection) error {
	if m.updateConnectionFn != nil {
		return m.updateConnectionFn(ctx, connection)
	}
	return nil
}

func (m *mockMapService) DeleteConnection(ctx context.Context, id int) error {
	if m.deleteConnectionFn != nil {
		return m.deleteConnectionFn(ctx, id)
	}
	return nil
}

func setupMapRouter(svc ServiceInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc)

	r.GET("/map/buildings", h.GetBuildings)
	r.GET("/map/buildings/:id", h.GetBuilding)
	r.GET("/map/rooms", h.GetRooms)
	r.GET("/map/rooms/:id", h.GetRoom)
	r.GET("/map/connections", h.GetConnections)
	r.GET("/map/connections/:id", h.GetConnection)
	r.GET("/map/path/:start/:end", h.GetPath)
	r.GET("/map/search", h.SearchRooms)
	r.POST("/admin/building", h.AddBuilding)
	r.PUT("/admin/building/:id", h.UpdateBuilding)
	r.DELETE("/admin/building/:id", h.DeleteBuilding)
	r.POST("/admin/room", h.AddRoom)
	r.PUT("/admin/room/:id", h.UpdateRoom)
	r.DELETE("/admin/room/:id", h.DeleteRoom)
	r.POST("/admin/connection", h.AddConnection)
	r.PUT("/admin/connection/:id", h.UpdateConnection)
	r.DELETE("/admin/connection/:id", h.DeleteConnection)

	return r
}

func TestGetBuildingsHandler(t *testing.T) {
	svc := &mockMapService{
		getAllBuildingsFn: func(context.Context) ([]Building, error) {
			return []Building{{ID: 1, Name: "Main"}}, nil
		},
	}
	r := setupMapRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/map/buildings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Main")

	svc.getAllBuildingsFn = func(context.Context) ([]Building, error) {
		return nil, errors.New("boom")
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPathHandler(t *testing.T) {
	svc := &mockMapService{
		findPathFn: func(_ context.Context, startRoom, endRoom string) ([]string, float64, error) {
			assert.Equal(t, "A", startRoom)
			assert.Equal(t, "C", endRoom)
			return []string{"A", "B", "C"}, 3, nil
		},
	}
	r := setupMapRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/map/path/A/C", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"dist\":3")

	svc.findPathFn = func(context.Context, string, string) ([]string, float64, error) {
		return nil, 0, errors.New("fail")
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSearchRoomsHandler(t *testing.T) {
	var gotBuildingID *int
	var gotFloor *int
	svc := &mockMapService{
		searchRoomsFn: func(_ context.Context, buildingID *int, floor *int) ([]Room, error) {
			gotBuildingID = buildingID
			gotFloor = floor
			return []Room{{ID: 1, Name: "101"}}, nil
		},
	}
	r := setupMapRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/map/search?building_id=2&floor=3", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	if assert.NotNil(t, gotBuildingID) {
		assert.Equal(t, 2, *gotBuildingID)
	}
	if assert.NotNil(t, gotFloor) {
		assert.Equal(t, 3, *gotFloor)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/map/search?building_id=x", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, badReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.searchRoomsFn = func(context.Context, *int, *int) ([]Room, error) {
		return nil, errors.New("db")
	}
	req = httptest.NewRequest(http.MethodGet, "/map/search", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoomHandlers(t *testing.T) {
	svc := &mockMapService{}
	r := setupMapRouter(svc)

	payload := AddRoomRequest{Name: "201", BuildingID: 1, Floor: 2}
	body, _ := json.Marshal(payload)

	svc.addRoomFn = func(context.Context, string, int, int) error { return nil }
	req := httptest.NewRequest(http.MethodPost, "/admin/room", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest(http.MethodPost, "/admin/room", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.updateRoomFn = func(context.Context, Room) error { return errors.New("update fail") }
	req = httptest.NewRequest(http.MethodPut, "/admin/room/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	req = httptest.NewRequest(http.MethodPut, "/admin/room/abc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.deleteRoomFn = func(context.Context, int) error { return nil }
	req = httptest.NewRequest(http.MethodDelete, "/admin/room/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestConnectionHandlers(t *testing.T) {
	svc := &mockMapService{}
	r := setupMapRouter(svc)

	payload := AddConnectionRequest{RoomFrom: "A", RoomTo: "B", Distance: 2.5, Type: "corridor"}
	body, _ := json.Marshal(payload)

	svc.addConnectionFn = func(context.Context, Connection) error { return nil }
	req := httptest.NewRequest(http.MethodPost, "/admin/connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	svc.updateConnectionFn = func(context.Context, Connection) error { return nil }
	req = httptest.NewRequest(http.MethodPut, "/admin/connection/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req = httptest.NewRequest(http.MethodPut, "/admin/connection/abc", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.deleteConnectionFn = func(context.Context, int) error { return errors.New("fail") }
	req = httptest.NewRequest(http.MethodDelete, "/admin/connection/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.deleteConnectionFn = func(context.Context, int) error { return nil }
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetEntityByIDHandlers(t *testing.T) {
	svc := &mockMapService{
		getBuildingByIDFn: func(_ context.Context, id int) (*Building, error) {
			assert.Equal(t, 1, id)
			return &Building{ID: 1, Name: "Main"}, nil
		},
		getRoomByIDFn: func(_ context.Context, id int) (*Room, error) {
			assert.Equal(t, 2, id)
			return &Room{ID: 2, Name: "101"}, nil
		},
		getConnectionByIDFn: func(_ context.Context, id int) (*Connection, error) {
			assert.Equal(t, 3, id)
			return &Connection{ID: 3, RoomFrom: "A", RoomTo: "B", Distance: 1}, nil
		},
	}
	r := setupMapRouter(svc)

	cases := []struct {
		url string
	}{
		{url: "/map/buildings/1"},
		{url: "/map/rooms/2"},
		{url: "/map/connections/3"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/map/rooms/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBuildingHandlers(t *testing.T) {
	svc := &mockMapService{}
	r := setupMapRouter(svc)

	payload := AddBuildingRequest{ID: 9, Name: "X", Address: "Addr"}
	body, _ := json.Marshal(payload)

	svc.addBuildingFn = func(context.Context, Building) error { return nil }
	req := httptest.NewRequest(http.MethodPost, "/admin/building", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	svc.updateBuildingFn = func(context.Context, Building) error { return nil }
	req = httptest.NewRequest(http.MethodPut, "/admin/building/9", bytes.NewBufferString(`{"name":"N","address":"A"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	svc.deleteBuildingFn = func(context.Context, int) error { return nil }
	req = httptest.NewRequest(http.MethodDelete, "/admin/building/9", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
