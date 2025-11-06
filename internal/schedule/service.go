package schedule

import (
	"context"
	"sort"
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

// GetScheduleByGroup возвращает расписание по названию группы
func (s *Service) GetScheduleByGroup(ctx context.Context, group string) ([]Lesson, error) {
	lessons, err := s.repo.GetLessonsByGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].DayOfWeek == lessons[j].DayOfWeek {
			return lessons[i].StartTime < lessons[j].StartTime
		}
		return lessons[i].DayOfWeek < lessons[j].DayOfWeek
	})

	return lessons, nil
}
