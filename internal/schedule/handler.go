package schedule

import (
	"fmt"
	"net/http"
	"strconv"
	"techup/internal/account"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ---------------- Lessons ----------------

// ListLessons godoc
// @Summary Get lessons by group and date range
// @Description Returns lessons for a group in a date range
// @Tags schedule
// @Produce json
// @Param group_id query int true "Group ID"
// @Param from query string true "From date YYYY-MM-DD"
// @Param to query string true "To date YYYY-MM-DD"
// @Success 200 {array} schedule.Lesson
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/schedule [get]
func (h *Handler) ListLessons(c *gin.Context) {
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

	lessons, err := h.service.ListLessonsByPeriod(c.Request.Context(), groupID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lessons)
}

// AddLesson godoc
// @Summary Add a new lesson
// @Description Creates a new lesson in the schedule
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param lesson body schedule.Lesson true "Lesson data"
// @Success 200 {object} schedule.Lesson
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/lesson [post]
func (h *Handler) AddLesson(c *gin.Context) {
	var l Lesson
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(l)
	err := h.service.AddLesson(c.Request.Context(), l)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

// UpdateLesson godoc
// @Summary Update lesson by ID
// @Description Updates an existing lesson
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param id path int true "Lesson ID"
// @Param lesson body schedule.Lesson true "Updated lesson data"
// @Success 200 {object} schedule.Lesson
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/lesson/{id} [put]
func (h *Handler) UpdateLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var l Lesson
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.ID = id
	if err := h.service.UpdateLesson(c.Request.Context(), l); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

// DeleteLesson godoc
// @Summary Delete lesson
// @Description Deletes a lesson by ID
// @Tags admin-schedule
// @Param id path int true "Lesson ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/lesson/{id} [delete]
func (h *Handler) DeleteLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteLesson(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- Faculties ----------------

// ListFaculties godoc
// @Summary Get all faculties
// @Description Returns all faculties
// @Tags schedule
// @Produce json
// @Success 200 {array} schedule.Faculty
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/schedule/faculties [get]
func (h *Handler) ListFaculties(c *gin.Context) {
	fac, err := h.service.ListFaculties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fac)
}

// AddFaculty godoc
// @Summary Add faculty
// @Description Adds a new faculty
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param faculty body schedule.Faculty true "Faculty data"
// @Success 200 {object} schedule.Faculty
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/faculty [post]
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
// @Summary Update faculty
// @Description Updates faculty by ID
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param id path int true "Faculty ID"
// @Param faculty body schedule.Faculty true "Updated faculty"
// @Success 200 {object} schedule.Faculty
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/faculty/{id} [put]
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
// @Summary Delete faculty
// @Description Deletes faculty by ID
// @Tags admin-schedule
// @Param id path int true "Faculty ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/faculty/{id} [delete]
func (h *Handler) DeleteFaculty(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteFaculty(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- Groups ----------------

// ListGroups godoc
// @Summary Get all groups
// @Description Returns list of groups
// @Tags schedule
// @Produce json
// @Success 200 {array} schedule.Group
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/schedule/groups [get]
func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.service.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// AddGroup godoc
// @Summary Add group
// @Description Adds a new group
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param group body schedule.Group true "Group data"
// @Success 200 {object} schedule.Group
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/group [post]
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
// @Summary Update group
// @Description Updates an existing group
// @Tags admin-schedule
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param group body schedule.Group true "Updated group"
// @Success 200 {object} schedule.Group
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/group/{id} [put]
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
// @Summary Delete group
// @Description Deletes group by ID
// @Tags admin-schedule
// @Param id path int true "Group ID"
// @Success 204 "No Content"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/admin/group/{id} [delete]
func (h *Handler) DeleteGroup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteGroup(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- Lesson Notes ----------------

// GetLessonNote godoc
// @Summary Get user note for lesson
// @Tags schedule
// @Produce json
// @Param id path int true "Lesson ID"
// @Success 200 {object} schedule.LessonNote
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/lessons/{id}/note [get]
func (h *Handler) GetLessonNote(c *gin.Context) {
	userID, err := account.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, _ := strconv.Atoi(c.Param("id"))

	note, err := h.service.GetLessonNote(c.Request.Context(), userID, lessonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if note == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, note)
}

// UpsertLessonNote godoc
// @Summary Create or update user note for lesson
// @Tags schedule
// @Accept json
// @Produce json
// @Param id path int true "Lesson ID"
// @Param body body map[string]string true "Note text"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/v1/lessons/{id}/note [post]
func (h *Handler) AddLessonNote(c *gin.Context) {
	userID, err := account.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	lessonID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpsertLessonNote(c.Request.Context(), userID, lessonID, req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
