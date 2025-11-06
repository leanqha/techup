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

func (r *Repository) GetLessonsByGroup(ctx context.Context, groupName string) ([]Lesson, error) {
	rows, err := r.db.Query(ctx,
		`SELECT l.id, l.group_id, l.day_of_week, l.start_time, l.end_time, l.subject, l.teacher, l.classroom, l.is_online, l.is_even_week
		 FROM lessons l
		 JOIN groups g ON g.id = l.group_id
		 WHERE g.name = $1
		 ORDER BY l.day_of_week, l.start_time`, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(&l.ID, &l.GroupName, &l.DayOfWeek, &l.StartTime, &l.EndTime,
			&l.Subject, &l.Teacher, &l.Classroom, &l.IsOnline, &l.IsEvenWeek); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}
