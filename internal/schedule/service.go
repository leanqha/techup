package schedule

import (
	"context"
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ---------- Faculties -----------

func (s *Service) AddFaculty(ctx context.Context, faculty Faculty) error {
	return s.repo.AddFaculty(ctx, faculty)
}

func (s *Service) ListFaculties(ctx context.Context) ([]Faculty, error) {
	return s.repo.ListFaculties(ctx)
}

func (s *Service) UpdateFaculty(ctx context.Context, faculty Faculty) error {
	return s.repo.UpdateFaculty(ctx, faculty)
}

func (s *Service) DeleteFaculty(ctx context.Context, id int) error {
	return s.repo.DeleteFaculty(ctx, id)
}

// Groups

func (s *Service) AddGroup(ctx context.Context, g Group) error {
	return s.repo.AddGroup(ctx, g)
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	return s.repo.ListGroups(ctx)
}

func (s *Service) UpdateGroup(ctx context.Context, g Group) error {
	return s.repo.UpdateGroup(ctx, g)
}

func (s *Service) DeleteGroup(ctx context.Context, id int) error {
	return s.repo.DeleteGroup(ctx, id)
}

// Lessons

func (s *Service) AddLesson(ctx context.Context, lesson Lesson) error {
	if lesson.StartTime.After(lesson.EndTime) {
		return errors.New("start_time must be before end_time")
	}
	return s.repo.AddLesson(ctx, lesson)
}

func (s *Service) UpdateLesson(ctx context.Context, lesson Lesson) error {
	if lesson.StartTime.After(lesson.EndTime) {
		return errors.New("start_time must be before end_time")
	}
	return s.repo.UpdateLesson(ctx, lesson)
}

func (s *Service) DeleteLesson(ctx context.Context, id int) error {
	return s.repo.DeleteLesson(ctx, id)
}

func (s *Service) GetLessons(
	ctx context.Context,
	groupID int,
	fromStr, toStr string,
) ([]Lesson, error) {
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, errors.New("invalid from date format")
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, errors.New("invalid to date format")
	}

	if from.After(to) {
		return nil, errors.New("from date must be before to date")
	}

	return s.repo.GetLessons(ctx, groupID, from, to)
}

// ---------- Lesson Notes ----------

func (s *Service) GetLessonNote(ctx context.Context, userID, lessonID int) (*LessonNote, error) {
	return s.repo.GetLessonNote(ctx, userID, lessonID)
}

func (s *Service) UpsertLessonNote(ctx context.Context, userID, lessonID int, text string) error {
	if len(text) > 5000 {
		return errors.New("note is too long")
	}

	return s.repo.UpsertLessonNote(ctx, userID, lessonID, text)
}
