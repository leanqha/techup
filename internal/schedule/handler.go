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

func (h *Handler) ListLessons(c *gin.Context) {
	lessons, err := h.service.ListLessons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, lessons)
}

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

func (h *Handler) DeleteLesson(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteLesson(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- Faculties ----------------

func (h *Handler) ListFaculties(c *gin.Context) {
	fac, err := h.service.ListFaculties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fac)
}

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

func (h *Handler) DeleteFaculty(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteFaculty(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- Groups ----------------

func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.service.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

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
