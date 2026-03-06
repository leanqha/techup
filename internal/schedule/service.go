package schedule

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"techup/internal/account"
	"techup/internal/apperrors"
	"time"

	"github.com/jackc/pgx/v5"
)

type RepositoryInterface interface {
	AddFaculty(ctx context.Context, faculty Faculty) error
	GetFaculty(ctx context.Context, id int) (*Faculty, error)
	ListFaculties(ctx context.Context) ([]Faculty, error)
	UpdateFaculty(ctx context.Context, faculty Faculty) error
	DeleteFaculty(ctx context.Context, id int) error
	AddGroup(ctx context.Context, g Group) error
	GetGroup(ctx context.Context, id int) (*Group, error)
	ListGroups(ctx context.Context) ([]Group, error)
	UpdateGroup(ctx context.Context, g Group) error
	DeleteGroup(ctx context.Context, id int) error
	AddLesson(ctx context.Context, lesson Lesson) error
	GetLesson(ctx context.Context, id int) (*LessonResponse, error)
	UpdateLesson(ctx context.Context, lesson Lesson) error
	DeleteLesson(ctx context.Context, id int) error
	GetLessons(ctx context.Context, groupID int, from, to time.Time) ([]LessonResponse, error)
	GetNote(ctx context.Context, note Note) (*Note, error)
	AddNote(ctx context.Context, note Note) error
	UpdateNote(ctx context.Context, note Note) error
	DeleteNote(ctx context.Context, note Note) error
	LessonExists(ctx context.Context, lessonID int) (bool, error)
	GetGroupIDByName(ctx context.Context, name string) (int, error)
	SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error)
	GetTeachers(ctx context.Context) ([]account.Account, error)
	GetClassrooms(ctx context.Context) ([]string, error)
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
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
		return apperrors.InvalidArgument("start_time must be before end_time")
	}
	return s.repo.AddLesson(ctx, lesson)
}

func (s *Service) UpdateLesson(ctx context.Context, lesson Lesson) error {
	if lesson.StartTime.After(lesson.EndTime) {
		return apperrors.InvalidArgument("start_time must be before end_time")
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
		return nil, apperrors.InvalidArgument("invalid from date format")
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, apperrors.InvalidArgument("invalid to date format")
	}

	if from.After(to) {
		return nil, apperrors.InvalidArgument("from date must be before to date")
	}

	return s.repo.GetLessons(ctx, groupID, from, to)
}

var ErrNoteTooLong = apperrors.InvalidArgument("note is too long")
var ErrLessonNotFound = apperrors.NotFound("lesson not found")
var ErrNoteNotFound = apperrors.NotFound("note not found")

func (s *Service) GetNote(ctx context.Context, note Note) (*Note, error) {
	exists, err := s.repo.LessonExists(ctx, note.LessonID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrLessonNotFound
	}

	return s.repo.GetNote(ctx, note)
}

func (s *Service) AddNote(ctx context.Context, note Note) error {
	if len(note.Text) > 5000 {
		return ErrNoteTooLong
	}

	exists, err := s.repo.LessonExists(ctx, note.LessonID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrLessonNotFound
	}

	return s.repo.AddNote(ctx, note)
}

func (s *Service) UpdateNote(ctx context.Context, note Note) error {
	if len(note.Text) > 5000 {
		return ErrNoteTooLong
	}

	exists, err := s.repo.LessonExists(ctx, note.LessonID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrLessonNotFound
	}

	err = s.repo.UpdateNote(ctx, note)
	if errors.Is(err, pgx.ErrNoRows) || apperrors.IsCode(err, apperrors.CodeNotFound) {
		return ErrNoteNotFound
	}
	return err
}

func (s *Service) DeleteNote(ctx context.Context, note Note) error {
	exists, err := s.repo.LessonExists(ctx, note.LessonID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrLessonNotFound
	}

	err = s.repo.DeleteNote(ctx, note)
	if errors.Is(err, pgx.ErrNoRows) || apperrors.IsCode(err, apperrors.CodeNotFound) {
		return ErrNoteNotFound
	}
	return err
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

		typeValue := rec[5]
		if typeValue == "" {
			return nil, apperrors.InvalidArgument(fmt.Sprintf("missing lesson type in row %d", i+1))
		}

		teacherID, err := strconv.Atoi(rec[6])
		if err != nil {
			return nil, apperrors.InvalidArgument(fmt.Sprintf("invalid teacher_id in row %d: %v", i+1, err))
		}

		res = append(res, LessonCSV{
			Date:      date,
			Group:     rec[1],
			StartTime: start,
			EndTime:   end,
			Subject:   rec[4],
			Type:      typeValue,
			TeacherID: teacherID,
			Classroom: rec[7],
		})
	}

	return res, nil
}

func (s *Service) ImportSchedule(ctx context.Context, lessons []LessonImport) error {
	for _, lesson := range lessons {
		groupName := strings.TrimSpace(lesson.GroupName)
		if groupName == "" {
			return apperrors.InvalidArgument("group name is required")
		}

		groupID, err := s.repo.GetGroupIDByName(ctx, groupName)
		if err != nil {
			if apperrors.IsCode(err, apperrors.CodeNotFound) {
				return apperrors.NotFound("group not found: " + groupName)
			}
			return err
		}

		l := Lesson{
			GroupID:   groupID,
			Date:      lesson.Date,
			StartTime: lesson.StartTime,
			EndTime:   lesson.EndTime,
			Subject:   lesson.Subject,
			Type:      lesson.Type,
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

func (s *Service) GetTeachers(ctx context.Context) ([]account.ProfileResponse, error) {
	teachers, err := s.repo.GetTeachers(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]account.ProfileResponse, 0, len(teachers))
	for _, acc := range teachers {
		res = append(res, account.ProfileResponse{
			ID:         acc.ID,
			UID:        acc.UID,
			Email:      acc.Email,
			FirstName:  acc.FirstName,
			MiddleName: acc.MiddleName,
			LastName:   acc.LastName,
			GroupID:    acc.GroupID,
			GroupName:  acc.GroupName,
			Role:       acc.Role,
		})
	}

	return res, nil
}

func (s *Service) GetClassrooms(ctx context.Context) ([]string, error) {
	return s.repo.GetClassrooms(ctx)
}

func (s *Service) GetFaculty(ctx context.Context, id int) (*Faculty, error) {
	return s.repo.GetFaculty(ctx, id)
}

func (s *Service) GetGroup(ctx context.Context, id int) (*Group, error) {
	return s.repo.GetGroup(ctx, id)
}

func (s *Service) GetLesson(ctx context.Context, id int) (*LessonResponse, error) {
	return s.repo.GetLesson(ctx, id)
}
