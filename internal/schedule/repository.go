package schedule

import (
	"context"
	"fmt"
	"techup/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ---------- FACULTIES ----------

func (r *Repository) AddFaculty(ctx context.Context, faculty Faculty) error {
	var id int
	query := `INSERT INTO faculties (name) VALUES ($1) RETURNING id`
	err := r.db.QueryRow(ctx, query, faculty.Name).Scan(&id)
	if err != nil {
		logger.LogSQLError(err, query, faculty.Name)
	}
	return err
}

func (r *Repository) GetFaculty(ctx context.Context, id int) (*Faculty, error) {
	var f Faculty
	query := `SELECT id, name FROM faculties WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(&f.ID, &f.Name)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &f, nil
}

func (r *Repository) ListFaculties(ctx context.Context) ([]Faculty, error) {
	query := `SELECT id, name FROM faculties ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "ListFaculties")
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

func (r *Repository) UpdateFaculty(ctx context.Context, faculty Faculty) error {
	query := `UPDATE faculties SET name=$1 WHERE id=$2`
	_, err := r.db.Exec(ctx, query, faculty.Name, faculty.ID)
	if err != nil {
		logger.LogSQLError(err, query, faculty.Name)
	}
	return err
}

func (r *Repository) DeleteFaculty(ctx context.Context, id int) error {
	query := `DELETE FROM faculties WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
	}
	return err
}

// ---------- GROUPS ----------

func (r *Repository) AddGroup(ctx context.Context, g Group) error {
	var id int
	query := `INSERT INTO groups (faculty_id, name, course, degree, year_start, specialization, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := r.db.QueryRow(ctx, query,
		g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive).Scan(&id)
	if err != nil {
		logger.LogSQLError(err, query,
			g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive)
	}
	return err
}

func (r *Repository) GetGroup(ctx context.Context, id int) (*Group, error) {
	var g Group
	query := `SELECT id, faculty_id, name, course, degree, year_start, specialization, is_active
		 FROM groups WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&g.ID, &g.FacultyID, &g.Name, &g.Course, &g.Degree, &g.YearStart, &g.Specialization, &g.IsActive)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &g, nil
}

func (r *Repository) ListGroupsByFaculty(ctx context.Context, facultyID int) ([]Group, error) {
	query := `SELECT id, faculty_id, name, course, degree, year_start, specialization, is_active
		 FROM groups WHERE faculty_id = $1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, facultyID)
	if err != nil {
		logger.LogSQLError(err, query, facultyID)
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

// ListGroups returns all groups without filtering
func (r *Repository) ListGroups(ctx context.Context) ([]Group, error) {
	query := `SELECT id, faculty_id, name, course, degree, year_start, specialization, is_active
		 FROM groups ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "ListGroups")
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

func (r *Repository) UpdateGroup(ctx context.Context, g Group) error {
	query := `UPDATE groups SET faculty_id=$1, name=$2, course=$3, degree=$4, year_start=$5, specialization=$6, is_active=$7 WHERE id=$8`
	_, err := r.db.Exec(ctx, query,
		g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive, g.ID)
	if err != nil {
		logger.LogSQLError(err, query, g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive, g.ID)
	}
	return err
}

func (r *Repository) DeleteGroup(ctx context.Context, id int) error {
	query := `DELETE FROM groups WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
	}
	return err
}

// ---------- LESSONS ----------

func (r *Repository) AddLesson(ctx context.Context, lesson Lesson) error {
	query := `INSERT INTO lessons (group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`
	err := r.db.QueryRow(ctx, query,
		lesson.GroupName, lesson.DayOfWeek, lesson.StartTime, lesson.EndTime, lesson.Subject,
		lesson.Teacher, lesson.Classroom, lesson.IsOnline, lesson.IsEvenWeek).Scan(&lesson.ID)
	if err != nil {
		logger.LogSQLError(err, query,
			lesson.GroupName, lesson.DayOfWeek, lesson.StartTime, lesson.EndTime, lesson.Subject,
			lesson.Teacher, lesson.Classroom, lesson.IsOnline, lesson.IsEvenWeek,
		)
	}
	return err
}

func (r *Repository) GetLesson(ctx context.Context, id int) (*Lesson, error) {
	var l Lesson
	query := `SELECT id, group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week, created_at
		 FROM lessons WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).
		Scan(&l.ID, &l.GroupName, &l.DayOfWeek, &l.StartTime, &l.EndTime, &l.Subject, &l.Teacher, &l.Classroom, &l.IsOnline, &l.IsEvenWeek, &l.CreatedAt)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return &l, nil
}

// ListLessons returns all lessons without filtering
func (r *Repository) ListLessons(ctx context.Context) ([]Lesson, error) {
	query := `SELECT id, group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week, created_at
		 FROM lessons ORDER BY day_of_week, start_time`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "ListLessons")
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

func (r *Repository) UpdateLesson(ctx context.Context, lesson Lesson) error {
	query := `UPDATE lessons SET group_name=$1, day_of_week=$2, start_time=$3, end_time=$4, subject=$5, teacher=$6, classroom=$7, is_online=$8, is_even_week=$9 WHERE id=$10`
	_, err := r.db.Exec(ctx, query,
		lesson.GroupName, lesson.DayOfWeek, lesson.StartTime, lesson.EndTime, lesson.Subject,
		lesson.Teacher, lesson.Classroom, lesson.IsOnline, lesson.IsEvenWeek, lesson.ID)
	if err != nil {
		logger.LogSQLError(err, query, lesson.GroupName, lesson.DayOfWeek, lesson.StartTime, lesson.EndTime, lesson.Subject,
			lesson.Teacher, lesson.Classroom, lesson.IsOnline, lesson.IsEvenWeek, lesson.ID,
		)
	}
	return err
}

func (r *Repository) DeleteLesson(ctx context.Context, id int) error {
	query := `DELETE FROM lessons WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
	}
	return err
}

// SearchLessons with optional filters: group, teacher, classroom, day_of_week, start/end time, is_even_week
func (r *Repository) SearchLessons(ctx context.Context, group, teacher, classroom, dayOfWeek, from, to string, isEvenWeek *bool) ([]Lesson, error) {
	query := `SELECT id, group_name, day_of_week, start_time, end_time, subject, teacher, classroom, is_online, is_even_week, created_at
	          FROM lessons WHERE 1=1`
	var args []interface{}
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
