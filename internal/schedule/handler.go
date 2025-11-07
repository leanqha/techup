package schedule

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// UploadSchedule godoc
// @Summary Upload schedule file
// @Description Upload a schedule file in PDF or Excel format
// @Tags schedule
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Schedule file (PDF or Excel)"
// @Success 200 {object} map[string]string "schedule uploaded and imported successfully"
// @Failure 400 {object} map[string]string "file is required or invalid file type"
// @Failure 500 {object} map[string]string "failed to save file or import lessons"
// @Router /schedule/upload [post]
func (h *Handler) UploadSchedule(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" && ext != ".xlsx" && ext != ".xls" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only PDF and Excel files are allowed"})
		return
	}

	savePath := "./" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	if ext == ".pdf" {
		if err := h.service.ImportScheduleFromPDF(c.Request.Context(), savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to import lessons from PDF"})
			return
		}
	} else {
		//TODO
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule uploaded and imported successfully"})
}

// AddLesson godoc
// @Summary Add a lesson
// @Description Add a new lesson to the schedule
// @Tags schedule
// @Accept json
// @Produce json
// @Param lesson body Lesson true "Lesson to add"
// @Success 200 {object} map[string]string "lesson added"
// @Failure 400 {object} map[string]string "invalid request"
// @Failure 500 {object} map[string]string "failed to add lesson"
// @Router /schedule/lesson [post]
func (h *Handler) AddLesson(c *gin.Context) {
	var req Lesson
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	lesson := Lesson{
		GroupName:  req.GroupName,
		DayOfWeek:  req.DayOfWeek,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Subject:    req.Subject,
		Teacher:    req.Teacher,
		Classroom:  req.Classroom,
		IsOnline:   req.IsOnline,
		IsEvenWeek: req.IsEvenWeek,
	}
	if err := h.service.AddLesson(c.Request.Context(), &lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "lesson added"})
}

// AddFaculty godoc
// @Summary Add a faculty
// @Description Add a new faculty
// @Tags schedule
// @Accept json
// @Produce json
// @Param faculty body Faculty true "Faculty to add"
// @Success 200 {object} map[string]string "faculty added"
// @Failure 400 {object} map[string]string "invalid request"
// @Failure 500 {object} map[string]string "failed to add faculty"
// @Router /schedule/faculty [post]
func (h *Handler) AddFaculty(c *gin.Context) {
	var req Faculty
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.AddFaculty(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"status": "faculty added"})

}

// AddGroup godoc
// @Summary Add a group
// @Description Add a new group
// @Tags schedule
// @Accept json
// @Produce json
// @Param group body Group true "Group to add"
// @Success 200 {object} map[string]string "group added"
// @Failure 400 {object} map[string]string "invalid request"
// @Failure 500 {object} map[string]string "failed to add group"
// @Router /schedule/group [post]
func (h *Handler) AddGroup(c *gin.Context) {
	var req Group
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.service.AddGroup(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"status": "group added"})

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
