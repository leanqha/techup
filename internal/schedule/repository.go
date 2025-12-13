package schedule

import (
	"context"
	"techup/internal/logger"
	"time"

	"github.com/jackc/pgx/v5"
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
	query := `
	INSERT INTO lessons (group_id, date, start_time, end_time, subject, teacher, classroom)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id`

	err := r.db.QueryRow(ctx, query,
		lesson.GroupID,
		lesson.Date,
		lesson.StartTime,
		lesson.EndTime,
		lesson.Subject,
		lesson.Teacher,
		lesson.Classroom,
	).Scan(&lesson.ID)

	if err != nil {
		logger.LogSQLError(err, query,
			lesson.GroupID,
			lesson.Date,
			lesson.StartTime,
			lesson.EndTime,
			lesson.Subject,
			lesson.Teacher,
			lesson.Classroom,
		)
	}

	return err
}

func (r *Repository) UpdateLesson(ctx context.Context, lesson Lesson) error {
	query := `
	UPDATE lessons
	SET group_id=$1, date=$2, start_time=$3, end_time=$4,
		subject=$5, teacher=$6, classroom=$7
	WHERE id=$8`

	_, err := r.db.Exec(ctx, query,
		lesson.GroupID,
		lesson.Date,
		lesson.StartTime,
		lesson.EndTime,
		lesson.Subject,
		lesson.Teacher,
		lesson.Classroom,
		lesson.ID,
	)

	if err != nil {
		logger.LogSQLError(err, query, lesson.ID)
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

func (r *Repository) ListLessonsByPeriod(
	ctx context.Context,
	groupID int,
	from, to time.Time,
) ([]Lesson, error) {

	query := `
	SELECT id, group_id, date, start_time, end_time,
		subject, teacher, classroom, created_at
	FROM lessons
	WHERE group_id=$1 AND date BETWEEN $2 AND $3
	ORDER BY date, start_time`

	rows, err := r.db.Query(ctx, query, groupID, from, to)
	if err != nil {
		logger.LogSQLError(err, query, groupID, from, to)
		return nil, err
	}
	defer rows.Close()

	var lessons []Lesson
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(
			&l.ID,
			&l.GroupID,
			&l.Date,
			&l.StartTime,
			&l.EndTime,
			&l.Subject,
			&l.Teacher,
			&l.Classroom,
			&l.CreatedAt,
		); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}

	return lessons, rows.Err()
}

// ---------- LESSON NOTES ----------

func (r *Repository) GetLessonNote(
	ctx context.Context,
	userID, lessonID int,
) (*LessonNote, error) {

	query := `
	SELECT id, user_id, lesson_id, text, created_at, updated_at
	FROM lesson_notes
	WHERE user_id=$1 AND lesson_id=$2`

	var n LessonNote
	err := r.db.QueryRow(ctx, query, userID, lessonID).
		Scan(&n.ID, &n.UserID, &n.LessonID, &n.Text, &n.CreatedAt, &n.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.LogSQLError(err, query, userID, lessonID)
	}

	return &n, err
}

func (r *Repository) UpsertLessonNote(
	ctx context.Context,
	userID, lessonID int,
	text string,
) error {

	query := `
	INSERT INTO lesson_notes (user_id, lesson_id, text)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, lesson_id)
	DO UPDATE SET text=EXCLUDED.text, updated_at=now()`

	_, err := r.db.Exec(ctx, query, userID, lessonID, text)
	if err != nil {
		logger.LogSQLError(err, query, userID, lessonID)
	}

	return err
}
