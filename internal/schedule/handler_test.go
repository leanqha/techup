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
	getLessonNoteFn  func(ctx context.Context, userID, lessonID int) (*LessonNote, error)
	addLessonNoteFn  func(ctx context.Context, userID, lessonID int, text string) error
	importScheduleFn func(ctx context.Context, lessons []Lesson) error
	searchLessonsFn  func(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error)
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

func (m *mockService) ListFaculties(ctx context.Context) ([]Faculty, error) {
	return nil, nil
}

func (m *mockService) AddFaculty(ctx context.Context, faculty Faculty) error {
	return nil
}

func (m *mockService) UpdateFaculty(ctx context.Context, faculty Faculty) error {
	return nil
}

func (m *mockService) DeleteFaculty(ctx context.Context, id int) error {
	return nil
}

func (m *mockService) ListGroups(ctx context.Context) ([]Group, error) {
	return nil, nil
}

func (m *mockService) AddGroup(ctx context.Context, g Group) error {
	return nil
}

func (m *mockService) UpdateGroup(ctx context.Context, g Group) error {
	return nil
}

func (m *mockService) DeleteGroup(ctx context.Context, id int) error {
	return nil
}

func (m *mockService) GetLessonNote(ctx context.Context, userID, lessonID int) (*LessonNote, error) {
	if m.getLessonNoteFn != nil {
		return m.getLessonNoteFn(ctx, userID, lessonID)
	}
	return nil, nil
}

func (m *mockService) AddLessonNote(ctx context.Context, userID, lessonID int, text string) error {
	if m.addLessonNoteFn != nil {
		return m.addLessonNoteFn(ctx, userID, lessonID, text)
	}
	return nil
}

func (m *mockService) ImportSchedule(ctx context.Context, lessons []Lesson) error {
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

func (m *mockService) GetTeachers(ctx context.Context) ([]account.Account, error) {
	return nil, nil
}

func (m *mockService) GetClassrooms(ctx context.Context) ([]string, error) {
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
	r.GET("/schedule/lessons/:id/note", h.GetLessonNote)
	r.POST("/schedule/lessons/:id/note", h.AddLessonNote)
	r.POST("/admin/schedule/import", h.ImportSchedule)
	r.GET("/schedule/search", h.SearchLessons)
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
		getLessonNoteFn: func(ctx context.Context, userID, lessonID int) (*LessonNote, error) {
			return nil, nil
		},
		addLessonNoteFn: func(ctx context.Context, userID, lessonID int, text string) error {
			if text == "fail" {
				return errors.New("save failed")
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
}

func TestImportScheduleHandler(t *testing.T) {
	svc := &mockService{}
	r := setupRouterWithService(svc, false)

	invalidJSON := httptest.NewRequest(http.MethodPost, "/admin/schedule/import", bytes.NewBufferString("{"))
	invalidJSON.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, invalidJSON)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	payload := []LessonRequest{{
		Group:     1,
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

	svc.importScheduleFn = func(ctx context.Context, lessons []Lesson) error {
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
