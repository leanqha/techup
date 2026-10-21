package schedule

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// errInvalidTime describes a malformed HH:MM time field in a slot request.
func errInvalidTime(field string) error {
	return fmt.Errorf("invalid %s format (expected HH:MM)", field)
}

// slotToResponse formats a stored slot for the API. Display names (group /
// teacher) are left for the client to resolve from its own caches; the list
// endpoint fills them via joins.
func slotToResponse(s ScheduleSlot) SlotResponse {
	return SlotResponse{
		ID:         s.ID,
		GroupID:    s.GroupID,
		Semester:   s.Semester,
		WeekParity: s.WeekParity,
		Weekday:    s.Weekday,
		StartTime:  s.StartTime.Format("15:04"),
		EndTime:    s.EndTime.Format("15:04"),
		Subject:    s.Subject,
		Type:       s.Type,
		TeacherID:  s.TeacherID,
		Classroom:  s.Classroom,
	}
}

// ---- Curriculum (group_subjects) ----

// ListGroupSubjects godoc
// @Summary      List curriculum entries
// @Description  Return curriculum entries (group semester plan), optionally filtered by group and semester.
// @Tags         Schedule
// @Produce      json
// @Param        group_id query int false "Group ID"
// @Param        semester query int false "Semester"
// @Success      200 {array} GroupSubjectResponse "Curriculum entries"
// @Failure      400 {object} map[string]string "Invalid filter value"
// @Failure      500 {object} map[string]string "Failed to load curriculum"
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
		sem, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid semester"})
			return
		}
		f.Semester = &sem
	}

	items, err := h.service.ListGroupSubjects(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// AddGroupSubject godoc
// @Summary      Add a curriculum entry (admin only)
// @Description  Add a subject to a group's semester plan, with an optional teacher.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        entry body GroupSubjectRequest true "Curriculum entry payload"
// @Success      201 {object} GroupSubject "Created curriculum entry"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      409 {object} map[string]string "Already in plan"
// @Failure      500 {object} map[string]string "Failed to create curriculum entry"
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
// @Summary      Update a curriculum entry (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id    path int true "Curriculum entry ID"
// @Param        entry body GroupSubjectRequest true "Updated payload"
// @Success      200 {object} GroupSubject "Updated curriculum entry"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      409 {object} map[string]string "Already in plan"
// @Failure      500 {object} map[string]string "Failed to update curriculum entry"
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
// @Summary      Delete a curriculum entry (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Curriculum entry ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      500 {object} map[string]string "Failed to delete curriculum entry"
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

// ---- Faculty exam sessions (faculty_sessions) ----

// facultySessionToResponse formats a stored session for the API. The faculty
// name is left empty here; the list endpoint fills it via a join.
func facultySessionToResponse(fs FacultySession) FacultySessionResponse {
	return FacultySessionResponse{
		ID:        fs.ID,
		FacultyID: fs.FacultyID,
		Name:      fs.Name,
		StartDate: fs.StartDate.Format("2006-01-02"),
		EndDate:   fs.EndDate.Format("2006-01-02"),
	}
}

// facultySessionFromRequest parses the YYYY-MM-DD dates of a request into a model.
func facultySessionFromRequest(req FacultySessionRequest) (FacultySession, error) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return FacultySession{}, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return FacultySession{}, fmt.Errorf("invalid end_date format (expected YYYY-MM-DD)")
	}
	return FacultySession{
		FacultyID: req.FacultyID,
		Name:      req.Name,
		StartDate: start,
		EndDate:   end,
	}, nil
}

// ListFacultySessions godoc
// @Summary      List faculty exam sessions
// @Description  Return faculty exam sessions (start/end dates), optionally filtered by faculty.
// @Tags         Schedule
// @Produce      json
// @Param        faculty_id query int false "Faculty ID"
// @Success      200 {array} FacultySessionResponse "Exam sessions"
// @Failure      400 {object} map[string]string "Invalid filter value"
// @Failure      500 {object} map[string]string "Failed to load sessions"
// @Router       /schedule/faculty-sessions [get]
func (h *Handler) ListFacultySessions(c *gin.Context) {
	var f FacultySessionFilter

	if v := c.Query("faculty_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid faculty_id"})
			return
		}
		f.FacultyID = &id
	}

	items, err := h.service.ListFacultySessions(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// AddFacultySession godoc
// @Summary      Create a faculty exam session (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        session body FacultySessionRequest true "Exam session payload"
// @Success      201 {object} FacultySessionResponse "Created session"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to create session"
// @Router       /admin/faculty-session [post]
func (h *Handler) AddFacultySession(c *gin.Context) {
	var req FacultySessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fs, err := facultySessionFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddFacultySession(c.Request.Context(), &fs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, facultySessionToResponse(fs))
}

// UpdateFacultySession godoc
// @Summary      Update a faculty exam session (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id      path int true "Exam session ID"
// @Param        session body FacultySessionRequest true "Updated payload"
// @Success      200 {object} FacultySessionResponse "Updated session"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      500 {object} map[string]string "Failed to update session"
// @Router       /admin/faculty-session/{id} [put]
func (h *Handler) UpdateFacultySession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req FacultySessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fs, err := facultySessionFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fs.ID = id

	if err := h.service.UpdateFacultySession(c.Request.Context(), fs); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, facultySessionToResponse(fs))
}

// DeleteFacultySession godoc
// @Summary      Delete a faculty exam session (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Exam session ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      500 {object} map[string]string "Failed to delete session"
// @Router       /admin/faculty-session/{id} [delete]
func (h *Handler) DeleteFacultySession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteFacultySession(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Schedule slots (schedule_slots) ----

// ListSlots godoc
// @Summary      List schedule slots
// @Description  Return schedule-builder slots, optionally filtered by group, semester or teacher. A teacher filter (across all groups) powers cross-group conflict detection.
// @Tags         Schedule
// @Produce      json
// @Param        group_id   query int false "Group ID"
// @Param        semester   query int false "Semester"
// @Param        teacher_id query int false "Teacher ID"
// @Success      200 {array} SlotResponse "Slots"
// @Failure      400 {object} map[string]string "Invalid filter value"
// @Failure      500 {object} map[string]string "Failed to load slots"
// @Router       /schedule/slots [get]
func (h *Handler) ListSlots(c *gin.Context) {
	var f SlotFilter

	if v := c.Query("group_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		f.GroupID = &id
	}
	if v := c.Query("semester"); v != "" {
		sem, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid semester"})
			return
		}
		f.Semester = &sem
	}
	if v := c.Query("teacher_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teacher_id"})
			return
		}
		f.TeacherID = &id
	}

	items, err := h.service.ListSlots(c.Request.Context(), f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// AddSlot godoc
// @Summary      Create a schedule slot (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        slot body SlotRequest true "Slot payload"
// @Success      201 {object} SlotResponse "Created slot"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to create slot"
// @Router       /admin/slot [post]
func (h *Handler) AddSlot(c *gin.Context) {
	var req SlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := slotFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddSlot(c.Request.Context(), &slot); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, slotToResponse(slot))
}

// UpdateSlot godoc
// @Summary      Update a schedule slot (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Slot ID"
// @Param        slot body SlotRequest true "Updated slot payload"
// @Success      200 {object} SlotResponse "Updated slot"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      500 {object} map[string]string "Failed to update slot"
// @Router       /admin/slot/{id} [put]
func (h *Handler) UpdateSlot(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req SlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := slotFromRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slot.ID = id

	if err := h.service.UpdateSlot(c.Request.Context(), slot); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, slotToResponse(slot))
}

// DeleteSlot godoc
// @Summary      Delete a schedule slot (admin only)
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Produce      json
// @Param        id path int true "Slot ID"
// @Success      204 {string} string "No Content"
// @Failure      400 {object} map[string]string "Invalid ID"
// @Failure      404 {object} map[string]string "Not found"
// @Failure      500 {object} map[string]string "Failed to delete slot"
// @Router       /admin/slot/{id} [delete]
func (h *Handler) DeleteSlot(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.DeleteSlot(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GenerateSchedule godoc
// @Summary      Generate lessons from slots (admin only)
// @Description  Expand a group's semester slots into concrete dated lessons across the given range, replacing that group's lessons in range.
// @Tags         Schedule
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        payload body GenerateScheduleRequest true "Generation payload"
// @Success      200 {object} GenerateScheduleResponse "Generation summary"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Failed to generate schedule"
// @Router       /admin/schedule/generate [post]
func (h *Handler) GenerateSchedule(c *gin.Context) {
	var req GenerateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
		return
	}

	res, err := h.service.GenerateSchedule(c.Request.Context(), req.GroupID, req.Semester, start, end, req.FirstWeekParity)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// slotFromRequest parses the HH:MM times of a slot request into a model.
func slotFromRequest(req SlotRequest) (ScheduleSlot, error) {
	start, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return ScheduleSlot{}, errInvalidTime("start_time")
	}
	end, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return ScheduleSlot{}, errInvalidTime("end_time")
	}
	return ScheduleSlot{
		GroupID:    req.GroupID,
		Semester:   req.Semester,
		WeekParity: req.WeekParity,
		Weekday:    req.Weekday,
		StartTime:  start,
		EndTime:    end,
		Subject:    req.Subject,
		Type:       req.Type,
		TeacherID:  req.TeacherID,
		Classroom:  req.Classroom,
	}, nil
}
