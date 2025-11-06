package schedule

import (
	"context"
	"sort"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ImportScheduleFromPDF — импорт расписания из PDF
func (s *Service) ImportScheduleFromPDF(ctx context.Context, path string) error {
	//text, err := parser.ParsePDF(path)
	//if err != nil {
	//	return err
	//}
	//
	//lessons, err := ParseLessons(text) // функция для распознавания предметов
	//for _, lesson := range lessons {
	//	if err := s.repo.SaveLesson(ctx, lesson); err != nil {
	//		return err
	//	}
	//}
	return nil
}

// GetScheduleByGroup возвращает расписание для указанной группы
func (s *Service) GetScheduleByGroup(ctx context.Context, group string) ([]Lesson, error) {
	lessons, err := s.repo.GetLessonsByGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	// Опционально сортируем по дню недели и времени начала
	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].DayOfWeek == lessons[j].DayOfWeek {
			return lessons[i].StartTime < lessons[j].StartTime
		}
		return lessons[i].DayOfWeek < lessons[j].DayOfWeek
	})

	return lessons, nil
}

func (s *Service) GetScheduleByProgram(ctx context.Context, program Program) ([]Lesson, error) {
	//TODO
	return nil, nil
}
