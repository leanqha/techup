package schedule

import (
	"context"
	"net/http"
	"strconv"
	"techup/internal/account"
	"techup/internal/apperrors"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	GetLessons(ctx context.Context, groupID int, fromStr, toStr string) ([]LessonResponse, error)
	GetLesson(ctx context.Context, id int) (*LessonResponse, error)
	AddLesson(ctx context.Context, lesson Lesson) error
	UpdateLesson(ctx context.Context, lesson Lesson) error
	DeleteLesson(ctx context.Context, id int) error
	ListFaculties(ctx context.Context) ([]Faculty, error)
	GetFaculty(ctx context.Context, id int) (*Faculty, error)
	AddFaculty(ctx context.Context, faculty Faculty) error
	UpdateFaculty(ctx context.Context, faculty Faculty) error
	DeleteFaculty(ctx context.Context, id int) error
	ListGroups(ctx context.Context) ([]Group, error)
	GetGroup(ctx context.Context, id int) (*Group, error)
	AddGroup(ctx context.Context, g Group) error
	UpdateGroup(ctx context.Context, g Group) error
	DeleteGroup(ctx context.Context, id int) error
	ListSubjects(ctx context.Context) ([]Subject, error)
	GetSubject(ctx context.Context, id int) (*Subject, error)
	AddSubject(ctx context.Context, subject *Subject) error
	UpdateSubject(ctx context.Context, subject Subject) error
	DeleteSubject(ctx context.Context, id int) error
	ListLessonTypes(ctx context.Context) ([]LessonType, error)
	GetLessonType(ctx context.Context, id int) (*LessonType, error)
	AddLessonType(ctx context.Context, lt *LessonType) error
	UpdateLessonType(ctx context.Context, lt LessonType) error
	DeleteLessonType(ctx context.Context, id int) error
	ListGroupSubjects(ctx context.Context, f GroupSubjectFilter) ([]GroupSubjectResponse, error)
	GetGroupSubject(ctx context.Context, id int) (*GroupSubjectResponse, error)
	AddGroupSubject(ctx context.Context, gs *GroupSubject) error
	UpdateGroupSubject(ctx context.Context, gs GroupSubject) error
	DeleteGroupSubject(ctx context.Context, id int) error
	GetNote(ctx context.Context, note Note) (*Note, error)
	AddNote(ctx context.Context, note Note) error
	UpdateNote(ctx context.Context, note Note) error
	DeleteNote(ctx context.Context, note Note) error
	ImportSchedule(ctx context.Context, lessons []LessonImport) error
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

func respondError(c *gin.Context, err error) {
	c.JSON(apperrors.StatusCode(err), gin.H{"error": apperrors.Message(err)})
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
		respondError(c, err)
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
		respondError(c, err)
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
		respondError(c, err)
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
		respondError(c, err)
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
		respondError(c, err)
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
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

// GetFaculty godoc
// @Summary      Get faculty by ID
// @Description  Return faculty by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Faculty ID"
// @Success      200 {object} Faculty "Faculty"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Faculty not found"
// @Failure      500 {object} map[string]string "Failed to load faculty"
// @Router       /schedule/faculties/{id} [get]
func (h *Handler) GetFaculty(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	faculty, err := h.service.GetFaculty(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, faculty)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var f Faculty
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f.ID = id
	if err := h.service.UpdateFaculty(c.Request.Context(), f); err != nil {
		respondError(c, err)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteFaculty(c.Request.Context(), id); err != nil {
		respondError(c, err)
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
		respondError(c, err)
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
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, g)
}

// GetGroup godoc
// @Summary      Get group by ID
// @Description  Return group by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Group ID"
// @Success      200 {object} Group "Group"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Group not found"
// @Failure      500 {object} map[string]string "Failed to load group"
// @Router       /schedule/groups/{id} [get]
func (h *Handler) GetGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	group, err := h.service.GetGroup(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, group)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var g Group
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if g.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group name is required"})
		return
	}

	g.ID = id
	err = h.service.UpdateGroup(c.Request.Context(), g)
	if err != nil {
		respondError(c, err)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteGroup(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListSubjects godoc
// @Summary      List subjects
// @Description  Return all subjects from the catalog, ordered by name.
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} Subject "Subjects list"
// @Failure      500 {object} map[string]string "Failed to load subjects"
// @Router       /schedule/subjects [get]
func (h *Handler) ListSubjects(c *gin.Context) {
	subjects, err := h.service.ListSubjects(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subjects)
}

// GetSubject godoc
// @Summary      Get subject by ID
// @Description  Return a subject by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Subject ID"
// @Success      200 {object} Subject "Subject"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Subject not found"
// @Failure      500 {object} map[string]string "Failed to load subject"
// @Router       /schedule/subjects/{id} [get]
func (h *Handler) GetSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	subject, err := h.service.GetSubject(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subject)
}

// AddSubject godoc
// @Summary      Create a subject (admin only)
// @Description  Add a new subject to the catalog.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        subject body SubjectRequest true "Subject payload"
// @Success      201 {object} Subject "Created subject"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      409 {object} map[string]string "Subject already exists"
// @Failure      500 {object} map[string]string "Failed to create subject"
// @Router       /admin/subject [post]
func (h *Handler) AddSubject(c *gin.Context) {
	var req SubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject := Subject{Name: req.Name}
	if err := h.service.AddSubject(c.Request.Context(), &subject); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, subject)
}

// UpdateSubject godoc
// @Summary      Update a subject (admin only)
// @Description  Update a subject's name by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id      path int true "Subject ID"
// @Param        subject body SubjectRequest true "Updated subject payload"
// @Success      200 {object} Subject "Updated subject"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Subject not found"
// @Failure      409 {object} map[string]string "Subject already exists"
// @Failure      500 {object} map[string]string "Failed to update subject"
// @Router       /admin/subject/{id} [put]
func (h *Handler) UpdateSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req SubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject := Subject{ID: id, Name: req.Name}
	if err := h.service.UpdateSubject(c.Request.Context(), subject); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, subject)
}

// DeleteSubject godoc
// @Summary      Delete a subject (admin only)
// @Description  Delete a subject from the catalog by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Subject ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Subject not found"
// @Failure      500 {object} map[string]string "Failed to delete subject"
// @Router       /admin/subject/{id} [delete]
func (h *Handler) DeleteSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteSubject(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListLessonTypes godoc
// @Summary      List lesson types
// @Description  Return all lesson types from the catalog (with their accent color).
// @Tags         Schedule
// @Produce      json
// @Success      200 {array} LessonType "Lesson types list"
// @Failure      500 {object} map[string]string "Failed to load lesson types"
// @Router       /schedule/lesson-types [get]
func (h *Handler) ListLessonTypes(c *gin.Context) {
	types, err := h.service.ListLessonTypes(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, types)
}

// GetLessonType godoc
// @Summary      Get lesson type by ID
// @Description  Return a lesson type by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Lesson type ID"
// @Success      200 {object} LessonType "Lesson type"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Lesson type not found"
// @Failure      500 {object} map[string]string "Failed to load lesson type"
// @Router       /schedule/lesson-types/{id} [get]
func (h *Handler) GetLessonType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	lt, err := h.service.GetLessonType(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, lt)
}

// AddLessonType godoc
// @Summary      Create a lesson type (admin only)
// @Description  Add a new lesson type to the catalog.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        lessonType body LessonTypeRequest true "Lesson type payload"
// @Success      201 {object} LessonType "Created lesson type"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      409 {object} map[string]string "Lesson type already exists"
// @Failure      500 {object} map[string]string "Failed to create lesson type"
// @Router       /admin/lesson-type [post]
func (h *Handler) AddLessonType(c *gin.Context) {
	var req LessonTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lt := LessonType{Name: req.Name, Color: req.Color}
	if err := h.service.AddLessonType(c.Request.Context(), &lt); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, lt)
}

// UpdateLessonType godoc
// @Summary      Update a lesson type (admin only)
// @Description  Update a lesson type's name and color by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id         path int true "Lesson type ID"
// @Param        lessonType body LessonTypeRequest true "Updated lesson type payload"
// @Success      200 {object} LessonType "Updated lesson type"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Lesson type not found"
// @Failure      409 {object} map[string]string "Lesson type already exists"
// @Failure      500 {object} map[string]string "Failed to update lesson type"
// @Router       /admin/lesson-type/{id} [put]
func (h *Handler) UpdateLessonType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req LessonTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lt := LessonType{ID: id, Name: req.Name, Color: req.Color}
	if err := h.service.UpdateLessonType(c.Request.Context(), lt); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, lt)
}

// DeleteLessonType godoc
// @Summary      Delete a lesson type (admin only)
// @Description  Delete a lesson type from the catalog by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Lesson type ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Lesson type not found"
// @Failure      500 {object} map[string]string "Failed to delete lesson type"
// @Router       /admin/lesson-type/{id} [delete]
func (h *Handler) DeleteLessonType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteLessonType(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListGroupSubjects godoc
// @Summary      List group semester-plan subjects
// @Description  Return subjects of groups' semester plans, optionally filtered by group and/or semester.
// @Tags         Schedule
// @Produce      json
// @Param        group_id query int false "Group ID"
// @Param        semester query int false "Semester number"
// @Success      200 {array} GroupSubjectResponse "Group subjects list"
// @Failure      400 {object} map[string]string "Invalid query parameters"
// @Failure      500 {object} map[string]string "Failed to load group subjects"
// @Router       /schedule/group-subjects [get]
func (h *Handler) ListGroupSubjects(c *gin.Context) {
	var f GroupSubjectFilter

	if v := c.Query("group_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		f.GroupID = &id
	}

	if v := c.Query("semester"); v != "" {
		semester, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid semester"})
			return
		}
		f.Semester = &semester
	}

	items, err := h.service.ListGroupSubjects(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetGroupSubject godoc
// @Summary      Get group semester-plan subject by ID
// @Description  Return a single group semester-plan entry by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Group subject ID"
// @Success      200 {object} GroupSubjectResponse "Group subject"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Group subject not found"
// @Failure      500 {object} map[string]string "Failed to load group subject"
// @Router       /schedule/group-subjects/{id} [get]
func (h *Handler) GetGroupSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.service.GetGroupSubject(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// AddGroupSubject godoc
// @Summary      Add a subject to a group's semester plan (admin only)
// @Description  Add a subject to a group for a given semester.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        groupSubject body GroupSubjectRequest true "Group subject payload"
// @Success      201 {object} GroupSubject "Created group subject"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      409 {object} map[string]string "Subject already in this group's semester plan"
// @Failure      500 {object} map[string]string "Failed to create group subject"
// @Router       /admin/group-subject [post]
func (h *Handler) AddGroupSubject(c *gin.Context) {
	var req GroupSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gs := GroupSubject{
		GroupID:   req.GroupID,
		SubjectID: req.SubjectID,
		Semester:  req.Semester,
		TeacherID: req.TeacherID,
	}
	if err := h.service.AddGroupSubject(c.Request.Context(), &gs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gs)
}

// UpdateGroupSubject godoc
// @Summary      Update a group's semester-plan subject (admin only)
// @Description  Update a group semester-plan entry by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id           path int true "Group subject ID"
// @Param        groupSubject body GroupSubjectRequest true "Updated group subject payload"
// @Success      200 {object} GroupSubject "Updated group subject"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Group subject not found"
// @Failure      409 {object} map[string]string "Subject already in this group's semester plan"
// @Failure      500 {object} map[string]string "Failed to update group subject"
// @Router       /admin/group-subject/{id} [put]
func (h *Handler) UpdateGroupSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req GroupSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gs := GroupSubject{
		ID:        id,
		GroupID:   req.GroupID,
		SubjectID: req.SubjectID,
		Semester:  req.Semester,
		TeacherID: req.TeacherID,
	}
	if err := h.service.UpdateGroupSubject(c.Request.Context(), gs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gs)
}

// DeleteGroupSubject godoc
// @Summary      Delete a group's semester-plan subject (admin only)
// @Description  Remove a subject from a group's semester plan by ID.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Group subject ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Group subject not found"
// @Failure      500 {object} map[string]string "Failed to delete group subject"
// @Router       /admin/group-subject/{id} [delete]
func (h *Handler) DeleteGroupSubject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteGroupSubject(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetNote godoc
// @Summary      Get lesson note
// @Description  Return the current user's note for a lesson. Returns 204 if no note exists.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Lesson ID"
// @Success      200 {object} Note "Lesson note"
// @Success      204 {string} string "No Content"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to load note"
// @Router       /schedule/lessons/{id}/note [get]
func (h *Handler) GetNote(c *gin.Context) {
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

	note, err := h.service.GetNote(c.Request.Context(), Note{UserID: userID, LessonID: lessonID})
	if err != nil {
		respondError(c, err)
		return
	}

	if note == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, note)
}

// AddNote godoc
// @Summary      Add lesson note
// @Description  Add or replace the current user's note for a lesson.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Lesson ID"
// @Param        note body NoteTextRequest true "Note payload (text)"
// @Success      200 {string} string "OK"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to save note"
// @Router       /schedule/lessons/{id}/note [post]
func (h *Handler) AddNote(c *gin.Context) {
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

	var req NoteTextRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note := Note{UserID: userID, LessonID: lessonID, Text: req.Text}
	if err := h.service.AddNote(c.Request.Context(), note); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// UpdateNote godoc
// @Summary      Update lesson note
// @Description  Update an existing note for the current user and lesson.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Lesson ID"
// @Param        note body NoteTextRequest true "Note payload (text)"
// @Success      200 {string} string "OK"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Lesson or note not found"
// @Failure      500 {object} map[string]string "Failed to update note"
// @Router       /schedule/lessons/{id}/note [put]
func (h *Handler) UpdateNote(c *gin.Context) {
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

	var req NoteTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateNote(c.Request.Context(), Note{UserID: userID, LessonID: lessonID, Text: req.Text}); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// DeleteNote godoc
// @Summary      Delete lesson note
// @Description  Delete the current user's note for a lesson.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Lesson ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Lesson or note not found"
// @Failure      500 {object} map[string]string "Failed to delete note"
// @Router       /schedule/lessons/{id}/note [delete]
func (h *Handler) DeleteNote(c *gin.Context) {
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

	note := Note{UserID: userID, LessonID: lessonID}
	if err := h.service.DeleteNote(c.Request.Context(), note); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ImportSchedule godoc
// @Summary      Import schedule (admin only)
// @Description  Import multiple lessons in bulk.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        lessons body []LessonImportRequest true "Lessons payload"
// @Success      201 {string} string "Created"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to import schedule"
// @Router       /admin/schedule/import [post]
func (h *Handler) ImportSchedule(c *gin.Context) {
	var dtos []LessonImportRequest
	if err := c.ShouldBindJSON(&dtos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json: " + err.Error()})
		return
	}

	lessons := make([]LessonImport, 0, len(dtos))
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

		lessons = append(lessons, LessonImport{
			GroupName: dto.Group,
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
		respondError(c, err)
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
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher_id"})
			return
		}
		f.TeacherID = &id
	}

	if v := c.Query("group_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
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
		respondError(c, err)
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
		c.JSON(apperrors.StatusCode(err), gin.H{"error": "failed to get teachers"})
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
		c.JSON(apperrors.StatusCode(err), gin.H{"error": "failed to get classrooms"})
		return
	}
	c.JSON(http.StatusOK, classrooms)
}

// GetLesson godoc
// @Summary      Get lesson by ID
// @Description  Return a single lesson by ID.
// @Tags         Schedule
// @Produce      json
// @Param        id path int true "Lesson ID"
// @Success      200 {object} LessonResponse "Lesson"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Lesson not found"
// @Failure      500 {object} map[string]string "Failed to load lesson"
// @Router       /schedule/lessons/{id} [get]
func (h *Handler) GetLesson(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	lesson, err := h.service.GetLesson(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, lesson)
}
