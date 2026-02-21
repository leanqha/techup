package schedule

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"techup/internal/account"
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
) ([]LessonResponse, error) {
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

		date, err := time.Parse("02.01.2006", rec[0])
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
			TeacherID: rec[5],
			Classroom: rec[6],
		})
	}

	return res, nil
}

func (s *Service) ImportSchedule(ctx context.Context, lessons []Lesson) error {
	for _, lesson := range lessons {
		// Получаем ID группы
		groupID, err := s.repo.GetGroupIDByName(ctx, strconv.Itoa(lesson.GroupID))
		if err != nil {
			return fmt.Errorf("group not found: %d", lesson.GroupID)
		}

		l := Lesson{
			GroupID:   groupID,
			Date:      lesson.Date,
			StartTime: lesson.StartTime,
			EndTime:   lesson.EndTime,
			Subject:   lesson.Subject,
			TeacherID: lesson.TeacherID,
			Classroom: lesson.Classroom,
		}

		if err := s.repo.AddLesson(ctx, l); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
	return s.repo.SearchLessons(ctx, f)
}

func (s *Service) GetTeachers(ctx context.Context) ([]account.Account, error) {
	return s.repo.GetTeachers(ctx)
}

func (s *Service) GetClassrooms(ctx context.Context) ([]string, error) {
	return s.repo.GetClassrooms(ctx)
}
