package schedule

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Faculties
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
	return s.repo.AddLesson(ctx, lesson)
}

func (s *Service) ListLessons(ctx context.Context) ([]Lesson, error) {
	return s.repo.ListLessons(ctx)
}

func (s *Service) UpdateLesson(ctx context.Context, lesson Lesson) error {
	return s.repo.UpdateLesson(ctx, lesson)
}

func (s *Service) DeleteLesson(ctx context.Context, id int) error {
	return s.repo.DeleteLesson(ctx, id)
}

// SearchSchedule performs search with extended filters
func (s *Service) SearchSchedule(ctx context.Context, group, teacher, classroom, dayOfWeek, from, to string, isEvenWeek *bool) ([]Lesson, error) {
	return s.repo.SearchLessons(ctx, group, teacher, classroom, dayOfWeek, from, to, isEvenWeek)
}
