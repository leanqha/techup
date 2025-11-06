package schedule

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveLesson(ctx context.Context, lesson *Lesson) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO lessons (id, program_id, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, group_number, is_even_week)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		lesson.ID,
		lesson.GroupNumber,
		lesson.ProgramID,
		lesson.Subject,
		lesson.Teacher,
		lesson.Classroom,
		lesson.StartTime,
		lesson.EndTime,
		lesson.DayOfWeek,
		lesson.IsOnline,
		lesson.IsEvenWeek,
	)
	return err
}

func (r *Repository) GetLessonsByGroup(ctx context.Context, group string) ([]Lesson, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, program_id, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, group_number, is_even_week
		 FROM lessons WHERE group_number = $1`, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		err := rows.Scan(
			&l.ID,
			&l.ProgramID,
			&l.DayOfWeek,
			&l.StartTime,
			&l.EndTime,
			&l.Subject,
			&l.Teacher,
			&l.Classroom,
			&l.IsOnline,
			&l.GroupNumber,
			&l.IsEvenWeek,
		)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *Repository) GetLessonsByProgram(ctx context.Context, programID int) ([]Lesson, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, group_number, program_id, subject, teacher, classroom, start_time, end_time, day_of_week, is_online, is_even_week
		 FROM lessons WHERE program_id = $1`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		err := rows.Scan(
			&l.ID,
			&l.GroupNumber,
			&l.ProgramID,
			&l.Subject,
			&l.Teacher,
			&l.Classroom,
			&l.StartTime,
			&l.EndTime,
			&l.DayOfWeek,
			&l.IsOnline,
			&l.IsEvenWeek,
		)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *Repository) DeleteLessonsByProgram(ctx context.Context, programID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM lessons WHERE program_id = $1`, programID)
	return err
}

func (r *Repository) GetFaculties(ctx context.Context) ([]Faculty, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM faculties`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var faculties []Faculty
	for rows.Next() {
		var f Faculty
		err := rows.Scan(&f.ID, &f.Name)
		if err != nil {
			return nil, err
		}
		faculties = append(faculties, f)
	}
	return faculties, rows.Err()
}

func (r *Repository) GetProgramsByFaculty(ctx context.Context, facultyID int) ([]Program, error) {
	rows, err := r.db.Query(ctx, `SELECT id, faculty_id, name FROM programs WHERE faculty_id = $1`, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var programs []Program
	for rows.Next() {
		var p Program
		err := rows.Scan(&p.ID, &p.FacultyID, &p.Name)
		if err != nil {
			return nil, err
		}
		programs = append(programs, p)
	}
	return programs, rows.Err()
}
