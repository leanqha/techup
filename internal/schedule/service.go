package schedule

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

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

func (s *Service) GetLessonNote(ctx context.Context, userID, lessonID int) (*LessonNote, error) {
	return s.repo.GetLessonNote(ctx, userID, lessonID)
}

func (s *Service) AddLessonNote(ctx context.Context, userID, lessonID int, text string) error {
	if len(text) > 5000 {
		return errors.New("note is too long")
	}

	return s.repo.AddLessonNote(ctx, userID, lessonID, text)
}

func (s *Service) parseCSV(r io.Reader) ([]LessonCSV, error) {
	reader := csv.NewReader(r)
	reader.Comma = ','

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var res []LessonCSV

	for i, rec := range records {
		if i == 0 {
			continue
		}

		date, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			return nil, err
		}

		start, err := time.Parse("15:04", rec[2])
		if err != nil {
			return nil, err
		}
		end, err := time.Parse("15:04", rec[3])
		if err != nil {
			return nil, err
		}

		res = append(res, LessonCSV{
			Date:      date,
			Group:     rec[1],
			StartTime: start,
			EndTime:   end,
			Subject:   rec[4],
			Teacher:   rec[5],
			Classroom: rec[6],
		})
	}

	return res, nil
}

func (s *Service) ImportSchedule(
	ctx context.Context,
	userID int,
	csvFile io.Reader,
) error {

	rows, err := s.parseCSV(csvFile)
	if err != nil {
		return err
	}

	for _, row := range rows {

		groupID, err := s.repo.GetGroupIDByName(ctx, row.Group)
		if err != nil {
			return err
		}

		lesson := Lesson{
			GroupID:   groupID,
			Date:      row.Date,
			StartTime: row.StartTime,
			EndTime:   row.EndTime,
			Subject:   row.Subject,
			Teacher:   row.Teacher,
			Classroom: row.Classroom,
		}

		if err := s.repo.AddLesson(ctx, lesson); err != nil {
			return err
		}
	}

	return nil
}
