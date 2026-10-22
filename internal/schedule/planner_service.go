package schedule

import (
	"context"
	"strings"
	"techup/internal/apperrors"
	"time"
)

// ---- Curriculum (group_subjects) ----

func (s *Service) ListGroupSubjects(ctx context.Context, f GroupSubjectFilter) ([]GroupSubjectResponse, error) {
	return s.repo.ListGroupSubjects(ctx, f)
}

func (s *Service) AddGroupSubject(ctx context.Context, gs *GroupSubject) error {
	if gs.Semester < 1 {
		return apperrors.InvalidArgument("semester must be a positive integer")
	}
	return s.repo.AddGroupSubject(ctx, gs)
}

func (s *Service) UpdateGroupSubject(ctx context.Context, gs GroupSubject) error {
	if gs.Semester < 1 {
		return apperrors.InvalidArgument("semester must be a positive integer")
	}
	return s.repo.UpdateGroupSubject(ctx, gs)
}

func (s *Service) DeleteGroupSubject(ctx context.Context, id int) error {
	return s.repo.DeleteGroupSubject(ctx, id)
}

// ---- Faculty exam sessions (faculty_sessions) ----

// normalizeFacultySession trims and validates an exam session before it is
// persisted.
func normalizeFacultySession(fs *FacultySession) error {
	if fs.EndDate.Before(fs.StartDate) {
		return apperrors.InvalidArgument("end_date must not be before start_date")
	}
	fs.Name = strings.TrimSpace(fs.Name)
	return nil
}

func (s *Service) ListFacultySessions(ctx context.Context, f FacultySessionFilter) ([]FacultySessionResponse, error) {
	return s.repo.ListFacultySessions(ctx, f)
}

func (s *Service) AddFacultySession(ctx context.Context, fs *FacultySession) error {
	if err := normalizeFacultySession(fs); err != nil {
		return err
	}
	return s.repo.AddFacultySession(ctx, fs)
}

func (s *Service) UpdateFacultySession(ctx context.Context, fs FacultySession) error {
	if err := normalizeFacultySession(&fs); err != nil {
		return err
	}
	return s.repo.UpdateFacultySession(ctx, fs)
}

func (s *Service) DeleteFacultySession(ctx context.Context, id int) error {
	return s.repo.DeleteFacultySession(ctx, id)
}

// ---- Schedule slots (schedule_slots) ----

// normalizeSlot trims, validates and defaults a slot before it is persisted.
func normalizeSlot(s *ScheduleSlot) error {
	s.WeekParity = strings.ToLower(strings.TrimSpace(s.WeekParity))
	if s.WeekParity != "odd" && s.WeekParity != "even" {
		return apperrors.InvalidArgument("week_parity must be 'odd' or 'even'")
	}
	if s.Weekday < 1 || s.Weekday > 7 {
		return apperrors.InvalidArgument("weekday must be between 1 and 7")
	}
	if !s.EndTime.After(s.StartTime) {
		return apperrors.InvalidArgument("start_time must be before end_time")
	}
	if s.Semester < 1 {
		return apperrors.InvalidArgument("semester must be a positive integer")
	}
	s.Subject = strings.TrimSpace(s.Subject)
	if s.Subject == "" {
		return apperrors.InvalidArgument("subject is required")
	}
	s.Type = strings.TrimSpace(s.Type)
	if s.Type == "" {
		s.Type = "Другое"
	}
	s.Classroom = strings.TrimSpace(s.Classroom)
	return nil
}

func (s *Service) ListSlots(ctx context.Context, f SlotFilter) ([]SlotResponse, error) {
	return s.repo.ListSlots(ctx, f)
}

func (s *Service) AddSlot(ctx context.Context, slot *ScheduleSlot) error {
	if err := normalizeSlot(slot); err != nil {
		return err
	}
	return s.repo.AddSlot(ctx, slot)
}

func (s *Service) UpdateSlot(ctx context.Context, slot ScheduleSlot) error {
	if err := normalizeSlot(&slot); err != nil {
		return err
	}
	return s.repo.UpdateSlot(ctx, slot)
}

func (s *Service) DeleteSlot(ctx context.Context, id int) error {
	return s.repo.DeleteSlot(ctx, id)
}

// ---- Generation (slots -> dated lessons) ----

// mondayOf returns midnight Monday of the ISO week containing t.
func mondayOf(t time.Time) time.Time {
	daysSinceMonday := (int(t.Weekday()) + 6) % 7 // Mon=0 .. Sun=6
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return d.AddDate(0, 0, -daysSinceMonday)
}

// otherParity flips "odd"/"even".
func otherParity(p string) string {
	if p == "odd" {
		return "even"
	}
	return "odd"
}

// expandSlots turns a two-week template into concrete lessons across
// [start, end] (inclusive). Week containing `start` is week 0 and has parity
// firstParity; parity then alternates every calendar week. Slots without a
// teacher are kept too — the resulting lesson simply has teacher_id NULL. Pure
// and deterministic.
func expandSlots(slots []ScheduleSlot, start, end time.Time, firstParity string) []Lesson {
	if firstParity != "even" {
		firstParity = "odd"
	}
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	monday := mondayOf(start)

	var lessons []Lesson
	for weekIndex := 0; ; weekIndex++ {
		weekMonday := monday.AddDate(0, 0, weekIndex*7)
		if weekMonday.After(end) {
			break
		}
		parity := firstParity
		if weekIndex%2 == 1 {
			parity = otherParity(firstParity)
		}
		for _, slot := range slots {
			if slot.WeekParity != parity {
				continue
			}
			date := weekMonday.AddDate(0, 0, slot.Weekday-1)
			if date.Before(start) || date.After(end) {
				continue
			}
			lessons = append(lessons, Lesson{
				GroupID:   slot.GroupID,
				Date:      date,
				StartTime: slot.StartTime,
				EndTime:   slot.EndTime,
				Subject:   slot.Subject,
				Type:      slot.Type,
				TeacherID: slot.TeacherID,
				Classroom: slot.Classroom,
			})
		}
	}
	return lessons
}

// GenerateSchedule expands a group's semester slots into concrete lessons for
// [start, end] and replaces that group's lessons in the same range. Slots
// without a teacher are expanded too (the lessons keep teacher_id NULL), so
// Skipped stays 0. Returns how many lessons were created.
func (s *Service) GenerateSchedule(ctx context.Context, groupID, semester int, start, end time.Time, firstParity string) (GenerateScheduleResponse, error) {
	var res GenerateScheduleResponse

	if end.Before(start) {
		return res, apperrors.InvalidArgument("end_date must not be before start_date")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return res, apperrors.InvalidArgument("date range is too large (max 1 year)")
	}

	slots, err := s.repo.ListSlotsForGeneration(ctx, groupID, semester)
	if err != nil {
		return res, err
	}

	lessons := expandSlots(slots, start, end, firstParity)
	if err := s.repo.ReplaceGroupLessonsInRange(ctx, groupID, start, end, lessons); err != nil {
		return res, err
	}

	res.Generated = len(lessons)
	return res, nil
}
