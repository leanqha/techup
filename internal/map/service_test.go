package maps

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMapRepo struct {
	getAllBuildingsFn   func(ctx context.Context) ([]Building, error)
	getBuildingByIDFn   func(ctx context.Context, id int) (*Building, error)
	addBuildingFn       func(ctx context.Context, building *Building) error
	updateBuildingFn    func(ctx context.Context, building *Building) error
	deleteBuildingFn    func(ctx context.Context, id int) error
	getRoomsFn          func(ctx context.Context) ([]Room, error)
	getRoomByIDFn       func(ctx context.Context, id int) (*Room, error)
	searchRoomsFn       func(ctx context.Context, room *Room, hasBuildingID, hasFloor bool) ([]Room, error)
	getConnectionsFn    func(ctx context.Context) ([]Connection, error)
	getConnectionByIDFn func(ctx context.Context, id int) (*Connection, error)
	addRoomFn           func(ctx context.Context, room *Room) error
	updateRoomFn        func(ctx context.Context, room *Room) error
	deleteRoomFn        func(ctx context.Context, id int) error
	addConnectionFn     func(ctx context.Context, conn *Connection) error
	updateConnectionFn  func(ctx context.Context, conn *Connection) error
	deleteConnectionFn  func(ctx context.Context, id int) error
}

func (m *mockMapRepo) GetAllBuildings(ctx context.Context) ([]Building, error) {
	if m.getAllBuildingsFn != nil {
		return m.getAllBuildingsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapRepo) GetBuildingByID(ctx context.Context, id int) (*Building, error) {
	if m.getBuildingByIDFn != nil {
		return m.getBuildingByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockMapRepo) AddBuilding(ctx context.Context, building *Building) error {
	if m.addBuildingFn != nil {
		return m.addBuildingFn(ctx, building)
	}
	return nil
}

func (m *mockMapRepo) UpdateBuilding(ctx context.Context, building *Building) error {
	if m.updateBuildingFn != nil {
		return m.updateBuildingFn(ctx, building)
	}
	return nil
}

func (m *mockMapRepo) DeleteBuilding(ctx context.Context, id int) error {
	if m.deleteBuildingFn != nil {
		return m.deleteBuildingFn(ctx, id)
	}
	return nil
}

func (m *mockMapRepo) GetRooms(ctx context.Context) ([]Room, error) {
	if m.getRoomsFn != nil {
		return m.getRoomsFn(ctx)
	}
	return nil, nil
}

func (m *mockMapRepo) GetRoomByID(ctx context.Context, id int) (*Room, error) {
	if m.getRoomByIDFn != nil {
		return m.getRoomByIDFn(ctx, id)
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

func (m *mockMapRepo) GetConnectionByID(ctx context.Context, id int) (*Connection, error) {
	if m.getConnectionByIDFn != nil {
		return m.getConnectionByIDFn(ctx, id)
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

func TestServiceGetByIDDelegatesToRepo(t *testing.T) {
	repo := &mockMapRepo{
		getBuildingByIDFn: func(_ context.Context, id int) (*Building, error) {
			assert.Equal(t, 1, id)
			return &Building{ID: id}, nil
		},
		getRoomByIDFn: func(_ context.Context, id int) (*Room, error) {
			assert.Equal(t, 2, id)
			return &Room{ID: id}, nil
		},
		getConnectionByIDFn: func(_ context.Context, id int) (*Connection, error) {
			assert.Equal(t, 3, id)
			return &Connection{ID: id}, nil
		},
	}
	svc := NewService(repo)

	building, err := svc.GetBuildingByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, building.ID)

	room, err := svc.GetRoomByID(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, room.ID)

	conn, err := svc.GetConnectionByID(context.Background(), 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, conn.ID)
}

func TestServiceBuildingCRUD_PassesValuesToRepo(t *testing.T) {
	var got Building
	repo := &mockMapRepo{
		addBuildingFn: func(_ context.Context, building *Building) error {
			got = *building
			return nil
		},
		updateBuildingFn: func(_ context.Context, building *Building) error {
			got = *building
			return nil
		},
		deleteBuildingFn: func(_ context.Context, id int) error {
			assert.Equal(t, 10, id)
			return nil
		},
	}
	svc := NewService(repo)

	err := svc.AddBuilding(context.Background(), Building{ID: 10, Name: "Main", Address: "Addr"})
	assert.NoError(t, err)
	assert.Equal(t, 10, got.ID)
	assert.Equal(t, "Main", got.Name)

	err = svc.UpdateBuilding(context.Background(), Building{ID: 10, Name: "New", Address: "NewAddr"})
	assert.NoError(t, err)
	assert.Equal(t, "New", got.Name)

	err = svc.DeleteBuilding(context.Background(), 10)
	assert.NoError(t, err)
}
