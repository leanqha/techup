package schedule

import (
	"fmt"
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

//func (h *Handler) GetScheduleByProgram(c *gin.Context) {
//	programID := c.Query("program_id")
//	if programID == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"error": "program_id query parameter is required"})
//		return
//	}
//
//	lessons, err := h.service.GetLessonsByProgram(programID)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get lessons"})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"lessons": lessons})
//}

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

	c.JSON(http.StatusOK, gin.H{"status": "schedule added"})
}

func (h *Handler) AddFaculty(c *gin.Context) {
	var req Faculty
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	fmt.Println(req)
	h.service.AddFaculty(c.Request.Context(), &req)
}

func (h *Handler) AddGroup(c *gin.Context) {
	var req Group
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	fmt.Println(req)
	h.service.AddGroup(c.Request.Context(), &req)
}
