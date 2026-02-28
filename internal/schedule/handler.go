package schedule

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"techup/internal/account"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	GetLessons(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error)
	AddLesson(ctx context.Context, lesson Lesson) error
	UpdateLesson(ctx context.Context, lesson Lesson) error
	DeleteLesson(ctx context.Context, id int) error
	ListFaculties(ctx context.Context) ([]Faculty, error)
	AddFaculty(ctx context.Context, faculty Faculty) error
	UpdateFaculty(ctx context.Context, faculty Faculty) error
	DeleteFaculty(ctx context.Context, id int) error
	ListGroups(ctx context.Context) ([]Group, error)
	AddGroup(ctx context.Context, g Group) error
	UpdateGroup(ctx context.Context, g Group) error
	DeleteGroup(ctx context.Context, id int) error
	GetLessonNote(ctx context.Context, userID, lessonID int) (*LessonNote, error)
	AddLessonNote(ctx context.Context, userID, lessonID int, text string) error
	ImportSchedule(ctx context.Context, lessons []Lesson) error
	SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error)
	GetTeachers(ctx context.Context) ([]account.ProfileResponse, error)
	GetClassrooms(ctx context.Context) ([]string, error)
}

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

// GetLessons godoc
// @Summary      List lessons for a group and date range
// @Description  Return lessons for a group between the provided dates (inclusive).
// @Tags         Schedule
// @Produce      json
// @Param        group_id query int true "Group ID"
// @Param        from     query string true "Start date (YYYY-MM-DD)"
// @Param        to       query string true "End date (YYYY-MM-DD)"
// @Success      200 {array} LessonResponse "Lessons list"
// @Failure      400 {object} map[string]string "Invalid query parameters"
// @Failure      500 {object} map[string]string "Failed to load lessons"
// @Router       /schedule/lessons [get]
func (h *Handler) GetLessons(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Query("group_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
		return
	}

	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to dates are required"})
		return
	}

	lessons, err := h.service.GetLessons(c.Request.Context(), groupID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lessons)
}

// AddLesson godoc
// @Summary      Create a lesson (admin only)
// @Description  Add a lesson for a group on a specific date/time.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        lesson body LessonRequest true "Lesson payload"
// @Success      201 {string} string "Created"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to create lesson"
// @Router       /admin/lesson [post]
func (h *Handler) AddLesson(c *gin.Context) {
	var req LessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	lesson := Lesson{
		GroupID:   req.Group,
		TeacherID: req.TeacherID,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Subject:   req.Subject,
		Type:      req.Type,
		Classroom: req.Classroom,
	}

	if err := h.service.AddLesson(c.Request.Context(), lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// UpdateLesson godoc
// @Summary      Update a lesson (admin only)
// @Description  Update lesson fields by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id     path int true "Lesson ID"
// @Param        lesson body LessonRequest true "Updated lesson payload"
// @Success      200 {string} string "OK"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to update lesson"
// @Router       /admin/lesson/{id} [put]
func (h *Handler) UpdateLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req LessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
		return
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format"})
		return
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format"})
		return
	}

	lesson := Lesson{
		ID:        id,
		GroupID:   req.Group,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Subject:   req.Subject,
		Type:      req.Type,
		TeacherID: req.TeacherID,
		Classroom: req.Classroom,
		CreatedAt: time.Time{},
	}

	if err := h.service.UpdateLesson(c.Request.Context(), lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteLesson godoc
// @Summary      Delete a lesson (admin only)
// @Description  Delete a lesson by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Lesson ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      500 {object} map[string]string "Failed to delete lesson"
// @Router       /admin/lesson/{id} [delete]
func (h *Handler) DeleteLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteLesson(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListFaculties godoc
// @Summary      List faculties
// @Description  Return all faculties.
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} Faculty "Faculties list"
// @Failure      500 {object} map[string]string "Failed to load faculties"
// @Router       /schedule/faculties [get]
func (h *Handler) ListFaculties(c *gin.Context) {
	fac, err := h.service.ListFaculties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fac)
}

// AddFaculty godoc
// @Summary      Create a faculty (admin only)
// @Description  Add a new faculty.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        faculty body Faculty true "Faculty payload"
// @Success      200 {object} Faculty "Created faculty"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to create faculty"
// @Router       /admin/faculty [post]
func (h *Handler) AddFaculty(c *gin.Context) {
	var f Faculty
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.AddFaculty(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, f)
}

// UpdateFaculty godoc
// @Summary      Update a faculty (admin only)
// @Description  Update a faculty by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id      path int true "Faculty ID"
// @Param        faculty body Faculty true "Updated faculty payload"
// @Success      200 {object} Faculty "Updated faculty"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to update faculty"
// @Router       /admin/faculty/{id} [put]
func (h *Handler) UpdateFaculty(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f Faculty
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f.ID = id
	if err := h.service.UpdateFaculty(c.Request.Context(), f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, f)
}

// DeleteFaculty godoc
// @Summary      Delete a faculty (admin only)
// @Description  Delete a faculty by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Faculty ID"
// @Success      204 {string} string "No Content"
// @Failure      500 {object} map[string]string "Failed to delete faculty"
// @Router       /admin/faculty/{id} [delete]
func (h *Handler) DeleteFaculty(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteFaculty(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListGroups godoc
// @Summary      List groups
// @Description  Return all groups.
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} Group "Groups list"
// @Failure      500 {object} map[string]string "Failed to load groups"
// @Router       /schedule/groups [get]
func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.service.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// AddGroup godoc
// @Summary      Create a group (admin only)
// @Description  Add a new group.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        group body Group true "Group payload"
// @Success      200 {object} Group "Created group"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to create group"
// @Router       /admin/group [post]
func (h *Handler) AddGroup(c *gin.Context) {
	var g Group
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.AddGroup(c.Request.Context(), g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

// UpdateGroup godoc
// @Summary      Update a group (admin only)
// @Description  Update group fields. Group name is required.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id    path int true "Group ID"
// @Param        group body Group true "Updated group payload"
// @Success      200 {integer} int "Updated group ID"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to update group"
// @Router       /admin/group/{id} [put]
func (h *Handler) UpdateGroup(c *gin.Context) {
	var g Group
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if g.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
		return
	}

	err := h.service.UpdateGroup(c.Request.Context(), g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, g.ID)
}

// DeleteGroup godoc
// @Summary      Delete a group (admin only)
// @Description  Delete a group by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Group ID"
// @Success      204 {string} string "No Content"
// @Failure      500 {object} map[string]string "Failed to delete group"
// @Router       /admin/group/{id} [delete]
func (h *Handler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteGroup(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetLessonNote godoc
// @Summary      Get lesson note
// @Description  Return the current user's note for a lesson. Returns 204 if no note exists.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Lesson ID"
// @Success      200 {object} LessonNote "Lesson note"
// @Success      204 {string} string "No Content"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to load note"
// @Router       /schedule/lessons/{id}/note [get]
func (h *Handler) GetLessonNote(c *gin.Context) {
	userID, err := account.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	note, err := h.service.GetLessonNote(c.Request.Context(), userID, lessonID)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if note == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, note)
}

// AddLessonNote godoc
// @Summary      Add lesson note
// @Description  Add or replace the current user's note for a lesson.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Lesson ID"
// @Param        note body map[string]string true "Note payload (text)"
// @Success      200 {string} string "OK"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to save note"
// @Router       /schedule/lessons/{id}/note [post]
func (h *Handler) AddLessonNote(c *gin.Context) {
	userID, err := account.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddLessonNote(c.Request.Context(), userID, lessonID, req.Text); err != nil {
		if errors.Is(err, ErrNoteTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrLessonNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ImportSchedule godoc
// @Summary      Import schedule (admin only)
// @Description  Import multiple lessons in bulk.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        lessons body []LessonRequest true "Lessons payload"
// @Success      201 {string} string "Created"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to import schedule"
// @Router       /admin/schedule/import [post]
func (h *Handler) ImportSchedule(c *gin.Context) {
	var dtos []LessonRequest
	if err := c.ShouldBindJSON(&dtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	lessons := make([]Lesson, 0, len(dtos))
	for _, dto := range dtos {
		date, err := time.Parse("2006-01-02", dto.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format: " + dto.Date})
			return
		}
		startTime, err := time.Parse("15:04", dto.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format: " + dto.StartTime})
			return
		}
		endTime, err := time.Parse("15:04", dto.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format: " + dto.EndTime})
			return
		}

		lessons = append(lessons, Lesson{
			GroupID:   dto.Group,
			TeacherID: dto.TeacherID,
			Date:      date,
			StartTime: startTime,
			EndTime:   endTime,
			Subject:   dto.Subject,
			Type:      dto.Type,
			Classroom: dto.Classroom,
		})
	}

	if err := h.service.ImportSchedule(c.Request.Context(), lessons); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// SearchLessons godoc
// @Summary      Search lessons
// @Description  Search lessons by optional filters.
// @Tags         Schedule
// @Produce      json
// @Param        date       query string false "Date (YYYY-MM-DD)"
// @Param        teacher_id query int false "Teacher ID"
// @Param        group_id   query int false "Group ID"
// @Param        classroom  query string false "Classroom"
// @Param        subject    query string false "Subject"
// @Success      200 {array} LessonResponse "Lessons list"
// @Failure      400 {object} map[string]string "Invalid filter value"
// @Failure      500 {object} map[string]string "Failed to search lessons"
// @Router       /schedule/search [get]
func (h *Handler) SearchLessons(c *gin.Context) {
	var f SearchLessonsFilter

	if v := c.Query("date"); v != "" {
		d, err := time.Parse("2006-01-02", v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
			return
		}
		f.Date = &d
	}

	if v := c.Query("teacher_id"); v != "" {
		id, _ := strconv.Atoi(v)
		f.TeacherID = &id
	}

	if v := c.Query("group_id"); v != "" {
		id, _ := strconv.Atoi(v)
		f.GroupID = &id
	}

	if v := c.Query("classroom"); v != "" {
		f.Classroom = &v
	}

	if v := c.Query("subject"); v != "" {
		f.Subject = &v
	}

	lessons, err := h.service.SearchLessons(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lessons)
}

// GetTeachers godoc
// @Summary      List teachers
// @Description  Return all teachers.
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} account.ProfileResponse "Teachers list"
// @Failure      500 {object} map[string]string "Failed to load teachers"
// @Router       /schedule/teachers [get]
func (h *Handler) GetTeachers(c *gin.Context) {
	teachers, err := h.service.GetTeachers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get teachers"})
		return
	}
	c.JSON(http.StatusOK, teachers)
}

// GetClassrooms godoc
// @Summary      List classrooms
// @Description  Return all available classrooms.
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} string "Classrooms list"
// @Failure      500 {object} map[string]string "Failed to load classrooms"
// @Router       /schedule/classrooms [get]
func (h *Handler) GetClassrooms(c *gin.Context) {
	classrooms, err := h.service.GetClassrooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get classrooms"})
		return
	}
	c.JSON(http.StatusOK, classrooms)
}
