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

// ImportScheduleFromPDF — заглушка для будущего импорта из PDF
func (s *Service) ImportScheduleFromPDF(ctx context.Context, path string) error {
	return nil
}

// AddFaculty добавляет новый факультет
func (s *Service) AddFaculty(ctx context.Context, faculty *Faculty) error {
	return s.repo.SaveFaculty(ctx, faculty)
}

// AddGroup добавляет новую группу
func (s *Service) AddGroup(ctx context.Context, group *Group) error {
	return s.repo.SaveGroup(ctx, group)
}

// AddLesson добавляет новое занятие
func (s *Service) AddLesson(ctx context.Context, lesson *Lesson) error {
	return s.repo.SaveLesson(ctx, lesson)
}

// GetFaculties возвращает список факультетов
func (s *Service) GetFaculties(ctx context.Context) ([]Faculty, error) {
	return s.repo.GetFaculties(ctx)
}

// GetGroupsByFaculty возвращает все группы факультета
func (s *Service) GetGroupsByFaculty(ctx context.Context, facultyID int) ([]Group, error) {
	return s.repo.GetGroupsByFaculty(ctx, facultyID)
}

// SearchSchedule performs search with extended filters: group, teacher, classroom, dayOfWeek, time range, isEvenWeek
func (s *Service) SearchSchedule(ctx context.Context, group, teacher, classroom, dayOfWeek, from, to string, isEvenWeek *bool) ([]Lesson, error) {
	return s.repo.SearchLessons(ctx, group, teacher, classroom, dayOfWeek, from, to, isEvenWeek)
}
