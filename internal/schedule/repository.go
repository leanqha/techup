package schedule

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ---------- FACULTIES ----------

func (r *Repository) SaveFaculty(ctx context.Context, faculty *Faculty) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO faculties (name) VALUES ($1)`,
		faculty.Name)
	return err
}

func (r *Repository) GetFaculties(ctx context.Context) ([]Faculty, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name FROM faculties ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var faculties []Faculty
	for rows.Next() {
		var f Faculty
		if err := rows.Scan(&f.ID, &f.Name); err != nil {
			return nil, err
		}
		faculties = append(faculties, f)
	}
	return faculties, rows.Err()
}

// ---------- GROUPS ----------

func (r *Repository) SaveGroup(ctx context.Context, group *Group) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO groups (faculty_id, name, course, degree, year_start, specialization, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		group.FacultyID, group.Name, group.Course, group.Degree, group.YearStart, group.Specialization, group.IsActive)
	return err
}

func (r *Repository) GetGroupsByFaculty(ctx context.Context, facultyID int) ([]Group, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, faculty_id, name, course, degree, year_start, specialization, is_active
		 FROM groups WHERE faculty_id = $1 ORDER BY name`, facultyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.FacultyID, &g.Name, &g.Course, &g.Degree, &g.YearStart, &g.Specialization, &g.IsActive); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// ---------- LESSONS ----------

func (r *Repository) SaveLesson(ctx context.Context, lesson *Lesson) error {
	query := `INSERT INTO lessons (group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 returning id
		 `
	err := r.db.QueryRow(ctx, query,
		lesson.GroupName, lesson.DayOfWeek, lesson.StartTime, lesson.EndTime, lesson.Subject,
		lesson.Teacher, lesson.Classroom, lesson.IsOnline, lesson.IsEvenWeek).Scan(&lesson.ID)
	if err != nil {
		return err
	}
	return nil
}

// SearchLessons with optional filters: group, teacher, classroom, day_of_week, start/end time, is_even_week
func (r *Repository) SearchLessons(ctx context.Context, group, teacher, classroom, dayOfWeek, from, to string, isEvenWeek *bool) ([]Lesson, error) {
	query := `SELECT id, group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week, created_at
	          FROM lessons WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if group != "" {
		query += fmt.Sprintf(" AND LOWER(group_name) = LOWER($%d)", argID)
		args = append(args, group)
		argID++
	}
	if teacher != "" {
		query += fmt.Sprintf(" AND LOWER(teacher) = LOWER($%d)", argID)
		args = append(args, teacher)
		argID++
	}
	if classroom != "" {
		query += fmt.Sprintf(" AND LOWER(classroom) = LOWER($%d)", argID)
		args = append(args, classroom)
		argID++
	}
	if dayOfWeek != "" {
		query += fmt.Sprintf(" AND LOWER(day_of_week) = LOWER($%d)", argID)
		args = append(args, dayOfWeek)
		argID++
	}
	if from != "" {
		query += fmt.Sprintf(" AND start_time >= $%d", argID)
		args = append(args, from)
		argID++
	}
	if to != "" {
		query += fmt.Sprintf(" AND end_time <= $%d", argID)
		args = append(args, to)
		argID++
	}
	if isEvenWeek != nil {
		query += fmt.Sprintf(" AND is_even_week = $%d", argID)
		args = append(args, *isEvenWeek)
		argID++
	}

	query += " ORDER BY day_of_week, start_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(
			&l.ID, &l.GroupName, &l.DayOfWeek,
			&l.StartTime, &l.EndTime, &l.Subject, &l.Teacher,
			&l.Classroom, &l.IsOnline, &l.IsEvenWeek, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, nil
}
