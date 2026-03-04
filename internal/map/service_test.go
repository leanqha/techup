package maps

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMapRepo struct {
	getAllBuildingsFn  func(ctx context.Context) ([]Building, error)
	searchRoomsFn      func(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error)
	getConnectionsFn   func(ctx context.Context) ([]Connection, error)
	addRoomFn          func(ctx context.Context, room *Room) error
	updateRoomFn       func(ctx context.Context, room *Room) error
	deleteRoomFn       func(ctx context.Context, id int) error
	addConnectionFn    func(ctx context.Context, conn *Connection) error
	updateConnectionFn func(ctx context.Context, conn *Connection) error
	deleteConnectionFn func(ctx context.Context, id int) error
}

func (m *mockMapRepo) GetAllBuildings(ctx context.Context) ([]Building, error) {
	if m.getAllBuildingsFn != nil {
		return m.getAllBuildingsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapRepo) SearchRooms(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error) {
	if m.searchRoomsFn != nil {
		return m.searchRoomsFn(ctx, room, hasBuildingID, hasFloor)
	}
	return nil, nil
}

func (m *mockMapRepo) GetConnections(ctx context.Context) ([]Connection, error) {
	if m.getConnectionsFn != nil {
		return m.getConnectionsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapRepo) AddRoom(ctx context.Context, room *Room) error {
	if m.addRoomFn != nil {
		return m.addRoomFn(ctx, room)
	}
	return nil
}

func (m *mockMapRepo) UpdateRoom(ctx context.Context, room *Room) error {
	if m.updateRoomFn != nil {
		return m.updateRoomFn(ctx, room)
	}
	return nil
}

func (m *mockMapRepo) DeleteRoom(ctx context.Context, id int) error {
	if m.deleteRoomFn != nil {
		return m.deleteRoomFn(ctx, id)
	}
	return nil
}

func (m *mockMapRepo) AddConnection(ctx context.Context, conn *Connection) error {
	if m.addConnectionFn != nil {
		return m.addConnectionFn(ctx, conn)
	}
	return nil
}

func (m *mockMapRepo) UpdateConnection(ctx context.Context, conn *Connection) error {
	if m.updateConnectionFn != nil {
		return m.updateConnectionFn(ctx, conn)
	}
	return nil
}

func (m *mockMapRepo) DeleteConnection(ctx context.Context, id int) error {
	if m.deleteConnectionFn != nil {
		return m.deleteConnectionFn(ctx, id)
	}
	return nil
}

func TestServiceSearchRooms_PropagatesFilters(t *testing.T) {
	var gotRoom Room
	var gotHasBuildingID bool
	var gotHasFloor bool

	repo := &mockMapRepo{
		searchRoomsFn: func(_ context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error) {
			gotRoom = *room
			gotHasBuildingID = hasBuildingID
			gotHasFloor = hasFloor
			return []Room{{ID: 1, Name: "101"}}, nil
		},
	}
	svc := NewService(repo)

	buildingID := 7
	floor := 2
	rooms, err := svc.SearchRooms(context.Background(), &buildingID, &floor)

	assert.NoError(t, err)
	assert.Len(t, rooms, 1)
	assert.Equal(t, 7, gotRoom.BuildingID)
	assert.Equal(t, 2, gotRoom.Floor)
	assert.True(t, gotHasBuildingID)
	assert.True(t, gotHasFloor)
}

func TestServiceFindPath_Success(t *testing.T) {
	repo := &mockMapRepo{
		getConnectionsFn: func(context.Context) ([]Connection, error) {
			return []Connection{
				{RoomFrom: "A", RoomTo: "B", Distance: 1},
				{RoomFrom: "B", RoomTo: "C", Distance: 2},
				{RoomFrom: "A", RoomTo: "C", Distance: 10},
			}, nil
		},
	}
	svc := NewService(repo)

	path, dist, err := svc.FindPath(context.Background(), "A", "C")

	assert.NoError(t, err)
	assert.Equal(t, []string{"A", "B", "C"}, path)
	assert.Equal(t, 3.0, dist)
}

func TestServiceFindPath_RepoError(t *testing.T) {
	repo := &mockMapRepo{
		getConnectionsFn: func(context.Context) ([]Connection, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewService(repo)

	_, _, err := svc.FindPath(context.Background(), "A", "C")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestServiceFindPath_MissingRoom(t *testing.T) {
	repo := &mockMapRepo{
		getConnectionsFn: func(context.Context) ([]Connection, error) {
			return []Connection{{RoomFrom: "A", RoomTo: "B", Distance: 1}}, nil
		},
	}
	svc := NewService(repo)

	_, _, err := svc.FindPath(context.Background(), "A", "Z")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end room Z not found")
}

func TestServiceCRUD_PassesValuesToRepo(t *testing.T) {
	var gotRoom Room
	var gotConnection Connection

	repo := &mockMapRepo{
		addRoomFn: func(_ context.Context, room *Room) error {
			gotRoom = *room
			return nil
		},
		addConnectionFn: func(_ context.Context, conn *Connection) error {
			gotConnection = *conn
			return nil
		},
	}
	svc := NewService(repo)

	err := svc.AddRoom(context.Background(), "201", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, "201", gotRoom.Name)
	assert.Equal(t, 1, gotRoom.BuildingID)
	assert.Equal(t, 2, gotRoom.Floor)

	err = svc.AddConnection(context.Background(), Connection{RoomFrom: "A", RoomTo: "B", Distance: 4.5, Type: "stairs"})
	assert.NoError(t, err)
	assert.Equal(t, "A", gotConnection.RoomFrom)
	assert.Equal(t, "B", gotConnection.RoomTo)
	assert.Equal(t, 4.5, gotConnection.Distance)
	assert.Equal(t, "stairs", gotConnection.Type)
}
