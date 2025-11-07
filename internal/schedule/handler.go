package schedule

import (
	"net/http"
	"path/filepath"
	"techup/internal/logger"

	"github.com/gin-gonic/gin"
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

// GetScheduleByGroup godoc
// @Summary Get schedule by group
// @Description Get lessons schedule for a specific group
// @Tags schedule
// @Accept json
// @Produce json
// @Param group query string true "Group name"
// @Success 200 {object} map[string]interface{} "lessons"
// @Failure 400 {object} map[string]string "group query parameter is required"
// @Failure 500 {object} map[string]string "failed to get lessons"
// @Router /schedule [get]
func (h *Handler) GetScheduleByGroup(c *gin.Context) {
	group := c.Query("group")
	if group == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group query parameter is required"})
		return
	}

	lessons, err := h.service.GetScheduleByGroup(c.Request.Context(), group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lessons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lessons": lessons})
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
		logger.Log.Warn().Err(err).Msg("failed to add lesson")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Log.Info().
		Int("id", lesson.ID).
		Msg("added lesson")
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
		logger.Log.Warn().Err(err).Msg("failed to add faculty")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	logger.Log.Info().
		Int("id", req.ID).
		Msg("added faculty")
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
		logger.Log.Warn().Err(err).Msg("failed to add group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	logger.Log.Info().
		Int("id", req.ID).
		Msg("added group")
	c.JSON(http.StatusOK, gin.H{"status": "group added"})

}
