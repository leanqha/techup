package schedule

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ---------------- Lessons ----------------

// ListLessons godoc
// @Summary Get all lessons
// @Description Returns a list of all lessons
// @Tags schedule
// @Produce json
// @Success 200 {array} schedule.Lesson
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/schedule/lessons [get]
func (h *Handler) ListLessons(c *gin.Context) {
	lessons, err := h.service.ListLessons(c.Request.Context())
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

// SearchSchedule godoc
// @Summary Search lessons schedule
// @Description Search lessons by group, teacher, classroom, day, time range, or even/odd week. At least one query parameter is required.
// @Tags schedule
// @Produce json
// @Param group query string false "Group name"
// @Param teacher query string false "Teacher name"
// @Param classroom query string false "Classroom"
// @Param day query string false "Day of the week"
// @Param from query string false "Start time in HH:MM format"
// @Param to query string false "End time in HH:MM format"
// @Param is_even_week query bool false "Even week (true/false)"
// @Success 200 {array} schedule.Lesson
// @Failure 400 {object} map[string]string "At least one query parameter is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /api/v1/schedule/search [get]
func (h *Handler) SearchSchedule(c *gin.Context) {
	group := c.Query("group")
	teacher := c.Query("teacher")
	classroom := c.Query("classroom")
	dayOfWeek := c.Query("day")
	from := c.Query("from")
	to := c.Query("to")

	var isEvenWeek *bool
	isEvenStr := c.Query("is_even_week")
	if isEvenStr != "" {
		val, err := strconv.ParseBool(isEvenStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "is_even_week must be a boolean"})
			return
		}
		isEvenWeek = &val
	}

	if group == "" && teacher == "" && classroom == "" && dayOfWeek == "" && from == "" && to == "" && isEvenWeek == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one query parameter is required"})
		return
	}

	lessons, err := h.service.SearchSchedule(c.Request.Context(), group, teacher, classroom, dayOfWeek, from, to, isEvenWeek)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lessons)
}
