package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"techup/internal/account"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	getLessonsFn     func(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error)
	addLessonFn      func(ctx context.Context, lesson Lesson) error
	updateLessonFn   func(ctx context.Context, lesson Lesson) error
	deleteLessonFn   func(ctx context.Context, id int) error
	getNoteFn        func(ctx context.Context, userID, lessonID int) (*Note, error)
	addNoteFn        func(ctx context.Context, userID, lessonID int, text string) error
	updateNoteFn     func(ctx context.Context, userID, lessonID int, text string) error
	deleteNoteFn     func(ctx context.Context, userID, lessonID int) error
	importScheduleFn func(ctx context.Context, lessons []LessonImport) error
	searchLessonsFn  func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error)
	getTeachersFn    func(ctx context.Context) ([]account.ProfileResponse, error)
	getClassroomsFn  func(ctx context.Context) ([]string, error)
}

func (m *mockService) GetLessons(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error) {
	if m.getLessonsFn != nil {
		return m.getLessonsFn(ctx, groupID, fromStr, toStr)
	}
	return nil, nil
}

func (m *mockService) AddLesson(ctx context.Context, lesson Lesson) error {
	if m.addLessonFn != nil {
		return m.addLessonFn(ctx, lesson)
	}
	return nil
}

func (m *mockService) UpdateLesson(ctx context.Context, lesson Lesson) error {
	if m.updateLessonFn != nil {
		return m.updateLessonFn(ctx, lesson)
	}
	return nil
}

func (m *mockService) DeleteLesson(ctx context.Context, id int) error {
	if m.deleteLessonFn != nil {
		return m.deleteLessonFn(ctx, id)
	}
	return nil
}

func (m *mockService) ListFaculties(_ context.Context) ([]Faculty, error) {
	return nil, nil
}

func (m *mockService) AddFaculty(_ context.Context, _ Faculty) error {
	return nil
}

func (m *mockService) UpdateFaculty(_ context.Context, _ Faculty) error {
	return nil
}

func (m *mockService) DeleteFaculty(_ context.Context, _ int) error {
	return nil
}

func (m *mockService) ListGroups(_ context.Context) ([]Group, error) {
	return nil, nil
}

func (m *mockService) AddGroup(_ context.Context, _ Group) error {
	return nil
}

func (m *mockService) UpdateGroup(_ context.Context, _ Group) error {
	return nil
}

func (m *mockService) DeleteGroup(_ context.Context, _ int) error {
	return nil
}

func (m *mockService) GetNote(ctx context.Context, userID, lessonID int) (*Note, error) {
	if m.getNoteFn != nil {
		return m.getNoteFn(ctx, userID, lessonID)
	}
	return nil, nil
}

func (m *mockService) AddNote(ctx context.Context, userID, lessonID int, text string) error {
	if m.addNoteFn != nil {
		return m.addNoteFn(ctx, userID, lessonID, text)
	}
	return nil
}

func (m *mockService) UpdateNote(ctx context.Context, userID, lessonID int, text string) error {
	if m.updateNoteFn != nil {
		return m.updateNoteFn(ctx, userID, lessonID, text)
	}
	return nil
}

func (m *mockService) DeleteNote(ctx context.Context, userID, lessonID int) error {
	if m.deleteNoteFn != nil {
		return m.deleteNoteFn(ctx, userID, lessonID)
	}
	return nil
}

func (m *mockService) ImportSchedule(ctx context.Context, lessons []LessonImport) error {
	if m.importScheduleFn != nil {
		return m.importScheduleFn(ctx, lessons)
	}
	return nil
}

func (m *mockService) SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
	if m.searchLessonsFn != nil {
		return m.searchLessonsFn(ctx, f)
	}
	return nil, nil
}

func (m *mockService) GetTeachers(ctx context.Context) ([]account.ProfileResponse, error) {
	if m.getTeachersFn != nil {
		return m.getTeachersFn(ctx)
	}
	return nil, nil
}

func (m *mockService) GetClassrooms(ctx context.Context) ([]string, error) {
	if m.getClassroomsFn != nil {
		return m.getClassroomsFn(ctx)
	}
	return nil, nil
}

func setupRouterWithService(svc ServiceInterface, withUser bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if withUser {
		r.Use(func(c *gin.Context) {
			c.Set("user_id", 1)
			c.Next()
		})
	}
	h := NewHandler(svc)
	r.GET("/schedule/lessons", h.GetLessons)
	r.POST("/admin/lesson", h.AddLesson)
	r.PUT("/admin/lesson/:id", h.UpdateLesson)
	r.DELETE("/admin/lesson/:id", h.DeleteLesson)
	r.GET("/schedule/lessons/:id/note", h.GetNote)
	r.POST("/schedule/lessons/:id/note", h.AddNote)
	r.PUT("/schedule/lessons/:id/note", h.UpdateNote)
	r.DELETE("/schedule/lessons/:id/note", h.DeleteNote)
	r.POST("/admin/schedule/import", h.ImportSchedule)
	r.GET("/schedule/search", h.SearchLessons)
	r.GET("/schedule/teachers", h.GetTeachers)
	r.GET("/schedule/classrooms", h.GetClassrooms)
	return r
}

func TestGetLessonsHandler(t *testing.T) {
	svc := &mockService{
		getLessonsFn: func(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error) {
			return []LessonResponse{{ID: 1, Subject: "Math"}}, nil
		},
	}
	r := setupRouterWithService(svc, false)

	badReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons?group_id=abc&from=2026-02-01&to=2026-02-02", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, badReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	missingReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons?group_id=1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, missingReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	okReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons?group_id=1&from=2026-02-01&to=2026-02-02", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, okReq)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Math")

	svc.getLessonsFn = func(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error) {
		return nil, errors.New("load failed")
	}
	failReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons?group_id=1&from=2026-02-01&to=2026-02-02", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, failReq)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAddAndUpdateLessonHandlers(t *testing.T) {
	svc := &mockService{}
	r := setupRouterWithService(svc, false)

	invalidJSON := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBufferString("{"))
	invalidJSON.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, invalidJSON)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload := LessonRequest{
		Group:     1,
		Date:      "2026-02-30",
		StartTime: "08:00",
		EndTime:   "09:00",
		Subject:   "Math",
		Type:      "lecture",
		Classroom: "101",
	}
	body, _ := json.Marshal(payload)
	invalidDate := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBuffer(body))
	invalidDate.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidDate)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload.Date = "2026-02-01"
	payload.StartTime = "xx"
	body, _ = json.Marshal(payload)
	invalidTime := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBuffer(body))
	invalidTime.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidTime)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload.StartTime = "08:00"
	payload.EndTime = "xx"
	body, _ = json.Marshal(payload)
	invalidEnd := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBuffer(body))
	invalidEnd.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidEnd)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.addLessonFn = func(ctx context.Context, lesson Lesson) error {
		return errors.New("create failed")
	}
	payload.EndTime = "09:00"
	body, _ = json.Marshal(payload)
	createFail := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBuffer(body))
	createFail.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, createFail)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.addLessonFn = nil
	successReq := httptest.NewRequest(http.MethodPost, "/admin/lesson", bytes.NewBuffer(body))
	successReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, successReq)
	assert.Equal(t, http.StatusCreated, w.Code)

	invalidID := httptest.NewRequest(http.MethodPut, "/admin/lesson/abc", bytes.NewBuffer(body))
	invalidID.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidID)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload.StartTime = "25:00"
	body, _ = json.Marshal(payload)
	invalidUpdateTime := httptest.NewRequest(http.MethodPut, "/admin/lesson/1", bytes.NewBuffer(body))
	invalidUpdateTime.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidUpdateTime)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.updateLessonFn = func(ctx context.Context, lesson Lesson) error {
		return errors.New("update failed")
	}
	payload.StartTime = "08:00"
	payload.EndTime = "09:00"
	body, _ = json.Marshal(payload)
	updateFail := httptest.NewRequest(http.MethodPut, "/admin/lesson/1", bytes.NewBuffer(body))
	updateFail.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateFail)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.updateLessonFn = nil
	updateOK := httptest.NewRequest(http.MethodPut, "/admin/lesson/1", bytes.NewBuffer(body))
	updateOK.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateOK)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteLessonHandler(t *testing.T) {
	svc := &mockService{}
	r := setupRouterWithService(svc, false)

	svc.deleteLessonFn = func(ctx context.Context, id int) error {
		return errors.New("delete failed")
	}
	failReq := httptest.NewRequest(http.MethodDelete, "/admin/lesson/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, failReq)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.deleteLessonFn = nil
	okReq := httptest.NewRequest(http.MethodDelete, "/admin/lesson/1", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, okReq)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestLessonNoteHandlers(t *testing.T) {
	svc := &mockService{
		getNoteFn: func(ctx context.Context, userID, lessonID int) (*Note, error) {
			return nil, nil
		},
		addNoteFn: func(ctx context.Context, userID, lessonID int, text string) error {
			if text == "fail" {
				return errors.New("save failed")
			}
			return nil
		},
		updateNoteFn: func(ctx context.Context, userID, lessonID int, text string) error {
			if text == "missing" {
				return ErrNoteNotFound
			}
			if text == "bad" {
				return errors.New("update failed")
			}
			return nil
		},
		deleteNoteFn: func(ctx context.Context, userID, lessonID int) error {
			if lessonID == 99 {
				return ErrNoteNotFound
			}
			if lessonID == 98 {
				return errors.New("delete failed")
			}
			return nil
		},
	}

	r := setupRouterWithService(svc, false)
	unauthReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons/1/note", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	r = setupRouterWithService(svc, true)
	noteReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons/1/note", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, noteReq)
	assert.Equal(t, http.StatusNoContent, w.Code)

	invalidBody := httptest.NewRequest(http.MethodPost, "/schedule/lessons/1/note", bytes.NewBufferString("{}"))
	invalidBody.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidBody)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload := map[string]string{"text": "fail"}
	body, _ := json.Marshal(payload)
	saveFail := httptest.NewRequest(http.MethodPost, "/schedule/lessons/1/note", bytes.NewBuffer(body))
	saveFail.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, saveFail)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	payload["text"] = "ok"
	body, _ = json.Marshal(payload)
	saveOK := httptest.NewRequest(http.MethodPost, "/schedule/lessons/1/note", bytes.NewBuffer(body))
	saveOK.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, saveOK)
	assert.Equal(t, http.StatusOK, w.Code)

	updateInvalid := httptest.NewRequest(http.MethodPut, "/schedule/lessons/1/note", bytes.NewBufferString("{}"))
	updateInvalid.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateInvalid)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload["text"] = "missing"
	body, _ = json.Marshal(payload)
	updateMissing := httptest.NewRequest(http.MethodPut, "/schedule/lessons/1/note", bytes.NewBuffer(body))
	updateMissing.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateMissing)
	assert.Equal(t, http.StatusNotFound, w.Code)

	payload["text"] = "ok"
	body, _ = json.Marshal(payload)
	updateOK := httptest.NewRequest(http.MethodPut, "/schedule/lessons/1/note", bytes.NewBuffer(body))
	updateOK.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, updateOK)
	assert.Equal(t, http.StatusOK, w.Code)

	deleteMissing := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/99/note", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, deleteMissing)
	assert.Equal(t, http.StatusNotFound, w.Code)

	deleteFail := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/98/note", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, deleteFail)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	deleteOK := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/1/note", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, deleteOK)
	assert.Equal(t, http.StatusNoContent, w.Code)

	invalidIDReq := httptest.NewRequest(http.MethodGet, "/schedule/lessons/abc/note", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidIDReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	invalidPostID := httptest.NewRequest(http.MethodPost, "/schedule/lessons/abc/note", bytes.NewBufferString("{}"))
	invalidPostID.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidPostID)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestImportScheduleHandler(t *testing.T) {
	svc := &mockService{}
	r := setupRouterWithService(svc, false)

	invalidJSON := httptest.NewRequest(http.MethodPost, "/admin/schedule/import", bytes.NewBufferString("{"))
	invalidJSON.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, invalidJSON)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload := []LessonImportRequest{{
		Group:     "G-1",
		Date:      "2026-02-30",
		StartTime: "08:00",
		EndTime:   "09:00",
		Subject:   "Math",
		Type:      "lecture",
		Classroom: "101",
	}}
	body, _ := json.Marshal(payload)
	invalidDate := httptest.NewRequest(http.MethodPost, "/admin/schedule/import", bytes.NewBuffer(body))
	invalidDate.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidDate)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	svc.importScheduleFn = func(ctx context.Context, lessons []LessonImport) error {
		return errors.New("import failed")
	}
	payload[0].Date = "2026-02-01"
	body, _ = json.Marshal(payload)
	importFail := httptest.NewRequest(http.MethodPost, "/admin/schedule/import", bytes.NewBuffer(body))
	importFail.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, importFail)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	svc.importScheduleFn = nil
	importOK := httptest.NewRequest(http.MethodPost, "/admin/schedule/import", bytes.NewBuffer(body))
	importOK.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, importOK)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSearchLessonsHandler(t *testing.T) {
	svc := &mockService{
		searchLessonsFn: func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
			return []LessonResponse{{ID: 1, Subject: "Math"}}, nil
		},
	}
	r := setupRouterWithService(svc, false)

	badDate := httptest.NewRequest(http.MethodGet, "/schedule/search?date=bad-date", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, badDate)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	okReq := httptest.NewRequest(http.MethodGet, "/schedule/search?date=2026-02-01", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, okReq)
	assert.Equal(t, http.StatusOK, w.Code)

	svc.searchLessonsFn = func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
		return nil, errors.New("search failed")
	}
	failReq := httptest.NewRequest(http.MethodGet, "/schedule/search?date=2026-02-01", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, failReq)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetTeachersHandler(t *testing.T) {
	svc := &mockService{
		getTeachersFn: func(ctx context.Context) ([]account.ProfileResponse, error) {
			return []account.ProfileResponse{{ID: 1, Email: "teacher@example.com", Role: "teacher"}}, nil
		},
	}
	r := setupRouterWithService(svc, false)

	okReq := httptest.NewRequest(http.MethodGet, "/schedule/teachers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, okReq)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "teacher@example.com")

	svc.getTeachersFn = func(ctx context.Context) ([]account.ProfileResponse, error) {
		return nil, errors.New("db down")
	}
	failReq := httptest.NewRequest(http.MethodGet, "/schedule/teachers", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, failReq)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get teachers")
}

func TestGetClassroomsHandler(t *testing.T) {
	svc := &mockService{
		getClassroomsFn: func(ctx context.Context) ([]string, error) {
			return []string{"101", "102"}, nil
		},
	}
	r := setupRouterWithService(svc, false)

	okReq := httptest.NewRequest(http.MethodGet, "/schedule/classrooms", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, okReq)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "101")

	svc.getClassroomsFn = func(ctx context.Context) ([]string, error) {
		return nil, errors.New("db down")
	}
	failReq := httptest.NewRequest(http.MethodGet, "/schedule/classrooms", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, failReq)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to get classrooms")
}
