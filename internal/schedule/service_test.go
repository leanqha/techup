package schedule

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"techup/internal/account"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	addLessonFn        func(ctx context.Context, lesson Lesson) error
	updateLessonFn     func(ctx context.Context, lesson Lesson) error
	getLessonsFn       func(ctx context.Context, groupID int, from, to time.Time) ([]LessonResponse, error)
	addNoteFn          func(ctx context.Context, userID, lessonID int, text string) error
	updateNoteFn       func(ctx context.Context, userID, lessonID int, text string) error
	deleteNoteFn       func(ctx context.Context, userID, lessonID int) error
	lessonExistsFn     func(ctx context.Context, lessonID int) (bool, error)
	getGroupIDByNameFn func(ctx context.Context, name string) (int, error)
	searchLessonsFn    func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error)
	getTeachersFn      func(ctx context.Context) ([]account.Account, error)
	getClassroomsFn    func(ctx context.Context) ([]string, error)
}

func (m *mockRepo) AddFaculty(_ context.Context, _ Faculty) error {
	return nil
}

func (m *mockRepo) ListFaculties(_ context.Context) ([]Faculty, error) {
	return nil, nil
}

func (m *mockRepo) UpdateFaculty(_ context.Context, _ Faculty) error {
	return nil
}

func (m *mockRepo) DeleteFaculty(_ context.Context, _ int) error {
	return nil
}

func (m *mockRepo) AddGroup(_ context.Context, _ Group) error {
	return nil
}

func (m *mockRepo) ListGroups(_ context.Context) ([]Group, error) {
	return nil, nil
}

func (m *mockRepo) UpdateGroup(_ context.Context, _ Group) error {
	return nil
}

func (m *mockRepo) DeleteGroup(_ context.Context, _ int) error {
	return nil
}

func (m *mockRepo) AddLesson(ctx context.Context, lesson Lesson) error {
	if m.addLessonFn != nil {
		return m.addLessonFn(ctx, lesson)
	}
	return nil
}

func (m *mockRepo) UpdateLesson(ctx context.Context, lesson Lesson) error {
	if m.updateLessonFn != nil {
		return m.updateLessonFn(ctx, lesson)
	}
	return nil
}

func (m *mockRepo) DeleteLesson(_ context.Context, _ int) error {
	return nil
}

func (m *mockRepo) GetLessons(ctx context.Context, groupID int, from, to time.Time) ([]LessonResponse, error) {
	if m.getLessonsFn != nil {
		return m.getLessonsFn(ctx, groupID, from, to)
	}
	return nil, nil
}

func (m *mockRepo) LessonExists(ctx context.Context, lessonID int) (bool, error) {
	if m.lessonExistsFn != nil {
		return m.lessonExistsFn(ctx, lessonID)
	}
	return true, nil
}

func (m *mockRepo) GetNote(_ context.Context, _, _ int) (*Note, error) {
	return nil, nil
}

func (m *mockRepo) AddNote(ctx context.Context, userID, lessonID int, text string) error {
	if m.addNoteFn != nil {
		return m.addNoteFn(ctx, userID, lessonID, text)
	}
	return nil
}

func (m *mockRepo) UpdateNote(ctx context.Context, userID, lessonID int, text string) error {
	if m.updateNoteFn != nil {
		return m.updateNoteFn(ctx, userID, lessonID, text)
	}
	return nil
}

func (m *mockRepo) DeleteNote(ctx context.Context, userID, lessonID int) error {
	if m.deleteNoteFn != nil {
		return m.deleteNoteFn(ctx, userID, lessonID)
	}
	return nil
}

func (m *mockRepo) GetGroupIDByName(ctx context.Context, name string) (int, error) {
	if m.getGroupIDByNameFn != nil {
		return m.getGroupIDByNameFn(ctx, name)
	}
	return 0, nil
}

func (m *mockRepo) SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
	if m.searchLessonsFn != nil {
		return m.searchLessonsFn(ctx, f)
	}
	return nil, nil
}

func (m *mockRepo) GetTeachers(ctx context.Context) ([]account.Account, error) {
	if m.getTeachersFn != nil {
		return m.getTeachersFn(ctx)
	}
	return nil, nil
}

func (m *mockRepo) GetClassrooms(ctx context.Context) ([]string, error) {
	if m.getClassroomsFn != nil {
		return m.getClassroomsFn(ctx)
	}
	return nil, nil
}

func TestServiceAddLessonValidation(t *testing.T) {
	called := false
	repo := &mockRepo{
		addLessonFn: func(ctx context.Context, lesson Lesson) error {
			called = true
			return nil
		},
	}
	service := NewService(repo)

	start := time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 27, 9, 0, 0, 0, time.UTC)
	err := service.AddLesson(context.Background(), Lesson{StartTime: start, EndTime: end})
	assert.Error(t, err)
	assert.False(t, called)

	called = false
	end = time.Date(2026, 2, 27, 11, 0, 0, 0, time.UTC)
	err = service.AddLesson(context.Background(), Lesson{StartTime: start, EndTime: end})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestServiceUpdateLessonValidation(t *testing.T) {
	called := false
	repo := &mockRepo{
		updateLessonFn: func(ctx context.Context, lesson Lesson) error {
			called = true
			return nil
		},
	}
	service := NewService(repo)

	start := time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 27, 11, 0, 0, 0, time.UTC)
	err := service.UpdateLesson(context.Background(), Lesson{StartTime: start, EndTime: end})
	assert.Error(t, err)
	assert.False(t, called)

	called = false
	end = time.Date(2026, 2, 27, 13, 0, 0, 0, time.UTC)
	err = service.UpdateLesson(context.Background(), Lesson{StartTime: start, EndTime: end})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestServiceGetLessonsDateValidation(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo)

	_, err := service.GetLessons(context.Background(), 1, "2026-02-30", "2026-03-01")
	assert.Error(t, err)

	_, err = service.GetLessons(context.Background(), 1, "2026-02-01", "2026-13-01")
	assert.Error(t, err)

	_, err = service.GetLessons(context.Background(), 1, "2026-03-01", "2026-02-01")
	assert.Error(t, err)
}

func TestServiceAddLessonNoteValidation(t *testing.T) {
	called := false
	repo := &mockRepo{
		addNoteFn: func(ctx context.Context, userID, lessonID int, text string) error {
			called = true
			return nil
		},
	}
	service := NewService(repo)

	text := strings.Repeat("a", 5001)
	err := service.AddNote(context.Background(), 1, 2, text)
	assert.Error(t, err)
	assert.False(t, called)
}

func TestServiceUpdateNoteValidationAndNotFound(t *testing.T) {
	called := false
	repo := &mockRepo{
		updateNoteFn: func(ctx context.Context, userID, lessonID int, text string) error {
			called = true
			return nil
		},
	}
	service := NewService(repo)

	text := strings.Repeat("a", 5001)
	err := service.UpdateNote(context.Background(), 1, 2, text)
	assert.ErrorIs(t, err, ErrNoteTooLong)
	assert.False(t, called)

	repo.updateNoteFn = func(ctx context.Context, userID, lessonID int, text string) error {
		return pgx.ErrNoRows
	}
	err = service.UpdateNote(context.Background(), 1, 2, "ok")
	assert.ErrorIs(t, err, ErrNoteNotFound)
}

func TestServiceDeleteNoteNotFound(t *testing.T) {
	repo := &mockRepo{
		deleteNoteFn: func(ctx context.Context, userID, lessonID int) error {
			return pgx.ErrNoRows
		},
	}
	service := NewService(repo)

	err := service.DeleteNote(context.Background(), 1, 2)
	assert.ErrorIs(t, err, ErrNoteNotFound)
}

func TestServiceImportSchedule(t *testing.T) {
	var captured Lesson
	repo := &mockRepo{
		getGroupIDByNameFn: func(ctx context.Context, name string) (int, error) {
			if name != "G-1" {
				return 0, errors.New("not found")
			}
			return 42, nil
		},
		addLessonFn: func(ctx context.Context, lesson Lesson) error {
			captured = lesson
			return nil
		},
	}
	service := NewService(repo)

	start := time.Date(2026, 2, 27, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC)
	lessons := []LessonImport{{
		GroupName: "G-1",
		Date:      start,
		StartTime: start,
		EndTime:   end,
		Subject:   "Math",
		Type:      "lecture",
		TeacherID: 7,
		Classroom: "101",
	}}

	err := service.ImportSchedule(context.Background(), lessons)
	assert.NoError(t, err)
	assert.Equal(t, 42, captured.GroupID)

	repo.getGroupIDByNameFn = func(ctx context.Context, name string) (int, error) {
		return 0, errors.New("not found")
	}
	err = service.ImportSchedule(context.Background(), lessons)
	if err == nil {
		t.Fatal("expected group not found error")
	}
	assert.Contains(t, err.Error(), "group not found")
}

func TestServiceParseCSV(t *testing.T) {
	repo := &mockRepo{}
	service := NewService(repo)

	input := strings.NewReader("Date,Group,Start,End,Subject,Type,TeacherID,Classroom\n01.02.2026,G-1,08:30,10:00,Math,lecture,12,101\n")
	items, err := service.parseCSV(input)
	assert.NoError(t, err)
	if assert.Len(t, items, 1) {
		assert.Equal(t, "G-1", items[0].Group)
		assert.Equal(t, 12, items[0].TeacherID)
		assert.Equal(t, "lecture", items[0].Type)
	}

	invalidType := strings.NewReader("Date,Group,Start,End,Subject,Type,TeacherID,Classroom\n01.02.2026,G-1,08:30,10:00,Math,,12,101\n")
	_, err = service.parseCSV(invalidType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing lesson type")
}

func TestServiceSearchLessons(t *testing.T) {
	var captured SearchLessonsFilter
	repo := &mockRepo{
		searchLessonsFn: func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
			captured = f
			return []LessonResponse{{ID: 1, Subject: "Math"}}, nil
		},
	}
	service := NewService(repo)

	teacherID := 10
	filter := SearchLessonsFilter{TeacherID: &teacherID}
	items, err := service.SearchLessons(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, filter, captured)
}

func TestServiceGetTeachersAndClassrooms(t *testing.T) {
	repo := &mockRepo{
		getTeachersFn: func(ctx context.Context) ([]account.Account, error) {
			return []account.Account{{ID: 1, Email: "t@example.com", Role: "teacher"}}, nil
		},
		getClassroomsFn: func(ctx context.Context) ([]string, error) {
			return []string{"101", "102"}, nil
		},
	}
	service := NewService(repo)

	teachers, err := service.GetTeachers(context.Background())
	assert.NoError(t, err)
	if assert.Len(t, teachers, 1) {
		assert.Equal(t, 1, teachers[0].ID)
		assert.Equal(t, "t@example.com", teachers[0].Email)
		assert.Equal(t, "teacher", teachers[0].Role)
	}

	rooms, err := service.GetClassrooms(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"101", "102"}, rooms)
}
