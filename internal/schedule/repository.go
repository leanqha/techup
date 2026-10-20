package schedule

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"techup/internal/account"
	"techup/internal/apperrors"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("faculty not found")
		}
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
	ct, err := r.db.Exec(ctx, query, faculty.Name, faculty.ID)
	if err != nil {
		logger.LogSQLError(err, query, faculty.Name)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("faculty not found")
	}
	return nil
}

func (r *Repository) DeleteFaculty(ctx context.Context, id int) error {
	query := `DELETE FROM faculties WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("faculty not found")
	}
	return nil
}

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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("group not found")
		}
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
	ct, err := r.db.Exec(ctx, query,
		g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive, g.ID)
	if err != nil {
		logger.LogSQLError(err, query, g.FacultyID, g.Name, g.Course, g.Degree, g.YearStart, g.Specialization, g.IsActive, g.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("group not found")
	}
	return nil
}

func (r *Repository) DeleteGroup(ctx context.Context, id int) error {
	query := `DELETE FROM groups WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("group not found")
	}
	return nil
}

func (r *Repository) AddSubject(ctx context.Context, subject *Subject) error {
	query := `INSERT INTO subjects (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, subject.Name).
		Scan(&subject.ID, &subject.CreatedAt, &subject.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already exists")
		}
		logger.LogSQLError(err, query, subject.Name)
		return err
	}
	return nil
}

func (r *Repository) GetSubject(ctx context.Context, id int) (*Subject, error) {
	var s Subject
	query := `SELECT id, name, created_at, updated_at FROM subjects WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		logger.LogSQLError(err, query, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("subject not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListSubjects(ctx context.Context) ([]Subject, error) {
	query := `SELECT id, name, created_at, updated_at FROM subjects ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "ListSubjects")
		return nil, err
	}
	defer rows.Close()

	var subjects []Subject
	for rows.Next() {
		var s Subject
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}
	return subjects, rows.Err()
}

func (r *Repository) UpdateSubject(ctx context.Context, subject Subject) error {
	query := `UPDATE subjects SET name=$1, updated_at=now() WHERE id=$2`
	ct, err := r.db.Exec(ctx, query, subject.Name, subject.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already exists")
		}
		logger.LogSQLError(err, query, subject.Name, subject.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("subject not found")
	}
	return nil
}

func (r *Repository) DeleteSubject(ctx context.Context, id int) error {
	query := `DELETE FROM subjects WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("subject not found")
	}
	return nil
}

func (r *Repository) AddLessonType(ctx context.Context, lt *LessonType) error {
	query := `INSERT INTO lesson_types (name, color) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, lt.Name, lt.Color).
		Scan(&lt.ID, &lt.CreatedAt, &lt.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("lesson type already exists")
		}
		logger.LogSQLError(err, query, lt.Name, lt.Color)
		return err
	}
	return nil
}

func (r *Repository) GetLessonType(ctx context.Context, id int) (*LessonType, error) {
	var lt LessonType
	query := `SELECT id, name, color, created_at, updated_at FROM lesson_types WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(&lt.ID, &lt.Name, &lt.Color, &lt.CreatedAt, &lt.UpdatedAt)
	if err != nil {
		logger.LogSQLError(err, query, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("lesson type not found")
		}
		return nil, err
	}
	return &lt, nil
}

func (r *Repository) ListLessonTypes(ctx context.Context) ([]LessonType, error) {
	query := `SELECT id, name, color, created_at, updated_at FROM lesson_types ORDER BY id`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.LogSQLError(err, query, "ListLessonTypes")
		return nil, err
	}
	defer rows.Close()

	var types []LessonType
	for rows.Next() {
		var lt LessonType
		if err := rows.Scan(&lt.ID, &lt.Name, &lt.Color, &lt.CreatedAt, &lt.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, lt)
	}
	return types, rows.Err()
}

func (r *Repository) UpdateLessonType(ctx context.Context, lt LessonType) error {
	query := `UPDATE lesson_types SET name=$1, color=$2, updated_at=now() WHERE id=$3`
	ct, err := r.db.Exec(ctx, query, lt.Name, lt.Color, lt.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("lesson type already exists")
		}
		logger.LogSQLError(err, query, lt.Name, lt.Color, lt.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("lesson type not found")
	}
	return nil
}

func (r *Repository) DeleteLessonType(ctx context.Context, id int) error {
	query := `DELETE FROM lesson_types WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("lesson type not found")
	}
	return nil
}

// groupSubjectSelect is the shared projection for group plan reads: it resolves
// the group, subject and teacher names so the API can return display-ready rows.
const groupSubjectSelect = `
	SELECT
		gs.id,
		gs.group_id,
		g.name,
		gs.subject_id,
		s.name,
		gs.semester,
		gs.teacher_id,
		a.first_name,
		a.middle_name,
		a.last_name
	FROM group_subjects gs
	LEFT JOIN groups g ON gs.group_id = g.id
	LEFT JOIN subjects s ON gs.subject_id = s.id
	LEFT JOIN accounts a ON gs.teacher_id = a.id
`

// scanGroupSubject reads one joined group-plan row and composes the teacher name.
func scanGroupSubject(row pgx.Row) (GroupSubjectResponse, error) {
	var (
		dto        GroupSubjectResponse
		firstName  *string
		middleName *string
		lastName   *string
	)

	if err := row.Scan(
		&dto.ID,
		&dto.GroupID,
		&dto.GroupName,
		&dto.SubjectID,
		&dto.SubjectName,
		&dto.Semester,
		&dto.TeacherID,
		&firstName,
		&middleName,
		&lastName,
	); err != nil {
		return GroupSubjectResponse{}, err
	}

	var parts []string
	if lastName != nil && *lastName != "" {
		parts = append(parts, *lastName)
	}
	if firstName != nil && *firstName != "" {
		parts = append(parts, *firstName)
	}
	if middleName != nil && *middleName != "" {
		parts = append(parts, *middleName)
	}
	dto.TeacherName = strings.Join(parts, " ")

	return dto, nil
}

func (r *Repository) AddGroupSubject(ctx context.Context, gs *GroupSubject) error {
	query := `INSERT INTO group_subjects (group_id, subject_id, semester, teacher_id)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(ctx, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID).
		Scan(&gs.ID, &gs.CreatedAt, &gs.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already in this group's semester plan")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return apperrors.InvalidArgument("group, subject or teacher not found")
		}
		logger.LogSQLError(err, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID)
		return err
	}
	return nil
}

func (r *Repository) GetGroupSubject(ctx context.Context, id int) (*GroupSubjectResponse, error) {
	query := groupSubjectSelect + " WHERE gs.id = $1"
	dto, err := scanGroupSubject(r.db.QueryRow(ctx, query, id))
	if err != nil {
		logger.LogSQLError(err, query, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("group subject not found")
		}
		return nil, err
	}
	return &dto, nil
}

func (r *Repository) ListGroupSubjects(ctx context.Context, f GroupSubjectFilter) ([]GroupSubjectResponse, error) {
	query := groupSubjectSelect

	var conditions []string
	var args []any
	i := 1

	if f.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("gs.group_id = $%d", i))
		args = append(args, *f.GroupID)
		i++
	}
	if f.Semester != nil {
		conditions = append(conditions, fmt.Sprintf("gs.semester = $%d", i))
		args = append(args, *f.Semester)
		i++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY gs.semester, s.name"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.LogSQLError(err, query, args...)
		return nil, err
	}
	defer rows.Close()

	var items []GroupSubjectResponse
	for rows.Next() {
		dto, err := scanGroupSubject(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, dto)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateGroupSubject(ctx context.Context, gs GroupSubject) error {
	query := `UPDATE group_subjects
		SET group_id=$1, subject_id=$2, semester=$3, teacher_id=$4, updated_at=now()
		WHERE id=$5`
	ct, err := r.db.Exec(ctx, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID, gs.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already in this group's semester plan")
		}
		if strings.Contains(err.Error(), "foreign key") {
			return apperrors.InvalidArgument("group, subject or teacher not found")
		}
		logger.LogSQLError(err, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID, gs.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("group subject not found")
	}
	return nil
}

func (r *Repository) DeleteGroupSubject(ctx context.Context, id int) error {
	query := `DELETE FROM group_subjects WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("group subject not found")
	}
	return nil
}

func (r *Repository) AddLesson(ctx context.Context, lesson Lesson) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	deleteQuery := `
	DELETE FROM lessons
	WHERE group_id = $1
	  AND date = $2
	  AND start_time < $4
	  AND end_time > $3
	`

	_, err = tx.Exec(ctx, deleteQuery,
		lesson.GroupID,
		lesson.Date,
		lesson.StartTime,
		lesson.EndTime,
	)
	if err != nil {
		return err
	}

	insertQuery := `
	INSERT INTO lessons (group_id, date, start_time, end_time, subject, type, teacher_id, classroom)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id
	`

	err = tx.QueryRow(ctx, insertQuery,
		lesson.GroupID,
		lesson.Date,
		lesson.StartTime,
		lesson.EndTime,
		lesson.Subject,
		lesson.Type,
		lesson.TeacherID,
		lesson.Classroom,
	).Scan(&lesson.ID)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) UpdateLesson(ctx context.Context, lesson Lesson) error {
	query := `
	UPDATE lessons
	SET group_id=$1, date=$2, start_time=$3, end_time=$4,
		subject=$5, type=$6, teacher_id=$7, classroom=$8
	WHERE id=$9`

	ct, err := r.db.Exec(ctx, query,
		lesson.GroupID,
		lesson.Date,
		lesson.StartTime,
		lesson.EndTime,
		lesson.Subject,
		lesson.Type,
		lesson.TeacherID,
		lesson.Classroom,
		lesson.ID,
	)

	if err != nil {
		logger.LogSQLError(err, query, lesson.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("lesson not found")
	}

	return nil
}

func (r *Repository) DeleteLesson(ctx context.Context, id int) error {
	query := `DELETE FROM lessons WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("lesson not found")
	}
	return nil
}

func (r *Repository) GetLessons(
	ctx context.Context,
	groupID int,
	from, to time.Time,
) ([]LessonResponse, error) {

	query := `
	SELECT 
		l.id,
		l.group_id,
		g.name,
		l.date,
		l.start_time,
		l.end_time,
		l.subject,
		l.type,
		l.teacher_id,
		a.first_name,
		a.middle_name,
		a.last_name,
		l.classroom
	FROM lessons l
	LEFT JOIN groups g ON l.group_id = g.id
	LEFT JOIN accounts a ON l.teacher_id = a.id
	WHERE l.group_id = $1 
		AND l.date BETWEEN $2 AND $3
	ORDER BY l.date, l.start_time
	`

	rows, err := r.db.Query(ctx, query, groupID, from, to)
	if err != nil {
		logger.LogSQLError(err, query, groupID, from, to)
		return nil, err
	}
	defer rows.Close()

	var lessons []LessonResponse

	for rows.Next() {
		var (
			dto        LessonResponse
			date       time.Time
			startTime  time.Time
			endTime    time.Time
			firstName  *string
			middleName *string
			lastName   *string
		)

		if err := rows.Scan(
			&dto.ID,
			&dto.Group.ID,
			&dto.Group.Name,
			&date,
			&startTime,
			&endTime,
			&dto.Subject,
			&dto.Type,
			&dto.Teacher.ID,
			&firstName,
			&middleName,
			&lastName,
			&dto.Classroom,
		); err != nil {
			return nil, err
		}

		dto.Date = date.Format("2006-01-02")
		dto.StartTime = startTime.Format("15:04")
		dto.EndTime = endTime.Format("15:04")

		var fullNameParts []string
		if lastName != nil && *lastName != "" {
			fullNameParts = append(fullNameParts, *lastName)
		}
		if firstName != nil && *firstName != "" {
			fullNameParts = append(fullNameParts, *firstName)
		}
		if middleName != nil && *middleName != "" {
			fullNameParts = append(fullNameParts, *middleName)
		}

		dto.Teacher.FullName = strings.Join(fullNameParts, " ")

		lessons = append(lessons, dto)
	}

	return lessons, rows.Err()
}

func (r *Repository) LessonExists(ctx context.Context, lessonID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM lessons WHERE id=$1)`
	err := r.db.QueryRow(ctx, query, lessonID).Scan(&exists)
	if err != nil {
		logger.LogSQLError(err, query, lessonID)
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetNote(
	ctx context.Context,
	note Note,
) (*Note, error) {

	query := `
	SELECT id, user_id, lesson_id, content, created_at, updated_at
	FROM notes
	WHERE user_id=$1 AND lesson_id=$2`

	var n Note
	err := r.db.QueryRow(ctx, query, note.UserID, note.LessonID).
		Scan(&n.ID, &n.UserID, &n.LessonID, &n.Text, &n.CreatedAt, &n.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		logger.LogSQLError(err, query, note.UserID, note.LessonID)
	}

	return &n, err
}

func (r *Repository) AddNote(
	ctx context.Context,
	note Note,
) error {

	query := `
	INSERT INTO notes (user_id, lesson_id, content)
	VALUES ($1, $2, $3)
	ON CONFLICT (user_id, lesson_id)
	DO UPDATE SET content=EXCLUDED.content, updated_at=now()`

	_, err := r.db.Exec(ctx, query, note.UserID, note.LessonID, note.Text)
	if err != nil {
		logger.LogSQLError(err, query, note.UserID, note.LessonID)
	}

	return err
}

func (r *Repository) UpdateNote(
	ctx context.Context,
	note Note,
) error {
	query := `
	UPDATE notes
	SET content = $3, updated_at = now()
	WHERE user_id = $1 AND lesson_id = $2`

	result, err := r.db.Exec(ctx, query, note.UserID, note.LessonID, note.Text)
	if err != nil {
		logger.LogSQLError(err, query, note.UserID, note.LessonID)
		return err
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("note not found")
	}

	return nil
}

func (r *Repository) DeleteNote(
	ctx context.Context,
	note Note,
) error {
	query := `DELETE FROM notes WHERE user_id = $1 AND lesson_id = $2`

	result, err := r.db.Exec(ctx, query, note.UserID, note.LessonID)
	if err != nil {
		logger.LogSQLError(err, query, note.UserID, note.LessonID)
		return err
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("note not found")
	}

	return nil
}

func (r *Repository) GetGroupIDByName(ctx context.Context, name string) (int, error) {

	var id int

	query := `
	SELECT id
	FROM groups
	WHERE name = $1 AND is_active = true
	`

	err := r.db.QueryRow(ctx, query, name).Scan(&id)
	if err != nil {
		logger.LogSQLError(err, query, name)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperrors.NotFound("group not found")
		}
		return 0, err
	}

	return id, nil
}

func (r *Repository) SearchLessons(ctx context.Context, f SearchLessonsFilter) ([]LessonResponse, error) {
	query := `
	SELECT 
		l.id,
		l.group_id,
		g.name,
		l.date,
		l.start_time,
		l.end_time,
		l.subject,
		l.type,
		l.teacher_id,
		a.first_name,
		a.middle_name,
		a.last_name,
		l.classroom
	FROM lessons l
	LEFT JOIN groups g ON l.group_id = g.id
	LEFT JOIN accounts a ON l.teacher_id = a.id
	`

	var conditions []string
	var args []any
	i := 1

	if f.Date != nil {
		conditions = append(conditions, fmt.Sprintf("l.date = $%d", i))
		args = append(args, *f.Date)
		i++
	}
	if f.TeacherID != nil {
		conditions = append(conditions, fmt.Sprintf("l.teacher_id = $%d", i))
		args = append(args, *f.TeacherID)
		i++
	}
	if f.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("l.group_id = $%d", i))
		args = append(args, *f.GroupID)
		i++
	}
	if f.Classroom != nil {
		conditions = append(conditions, fmt.Sprintf("l.classroom = $%d", i))
		args = append(args, *f.Classroom)
		i++
	}
	if f.Subject != nil {
		conditions = append(conditions, fmt.Sprintf("l.subject ILIKE $%d", i))
		args = append(args, "%"+*f.Subject+"%")
		i++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY l.date, l.start_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []LessonResponse

	for rows.Next() {
		var (
			dto        LessonResponse
			date       time.Time
			startTime  time.Time
			endTime    time.Time
			firstName  *string
			middleName *string
			lastName   *string
		)

		if err := rows.Scan(
			&dto.ID,
			&dto.Group.ID,
			&dto.Group.Name,
			&date,
			&startTime,
			&endTime,
			&dto.Subject,
			&dto.Type,
			&dto.Teacher.ID,
			&firstName,
			&middleName,
			&lastName,
			&dto.Classroom,
		); err != nil {
			return nil, err
		}

		dto.Date = date.Format("2006-01-02")
		dto.StartTime = startTime.Format("15:04")
		dto.EndTime = endTime.Format("15:04")

		var parts []string
		if lastName != nil && *lastName != "" {
			parts = append(parts, *lastName)
		}
		if firstName != nil && *firstName != "" {
			parts = append(parts, *firstName)
		}
		if middleName != nil && *middleName != "" {
			parts = append(parts, *middleName)
		}
		dto.Teacher.FullName = strings.Join(parts, " ")

		lessons = append(lessons, dto)
	}

	return lessons, rows.Err()
}

func (r *Repository) GetTeachers(ctx context.Context) ([]account.Account, error) {
	query := `SELECT id, uid, email, first_name, middle_name, last_name, role, created_at 
                                        FROM accounts 
                                        WHERE role = 'teacher'`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []account.Account
	for rows.Next() {
		var acc account.Account
		if err := rows.Scan(&acc.ID, &acc.UID, &acc.Email, &acc.FirstName, &acc.MiddleName, &acc.LastName, &acc.Role, &acc.CreatedAt); err != nil {
			return nil, err
		}
		teachers = append(teachers, acc)
	}

	return teachers, nil
}

func (r *Repository) GetClassrooms(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT name FROM rooms`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classrooms []string
	for rows.Next() {
		var room string
		if err := rows.Scan(&room); err != nil {
			return nil, err
		}
		classrooms = append(classrooms, room)
	}

	return classrooms, nil
}

func (r *Repository) GetLesson(ctx context.Context, id int) (*LessonResponse, error) {
	query := `
	SELECT
		l.id,
		l.group_id,
		g.name,
		l.date,
		l.start_time,
		l.end_time,
		l.subject,
		l.type,
		l.teacher_id,
		a.first_name,
		a.middle_name,
		a.last_name,
		l.classroom
	FROM lessons l
	LEFT JOIN groups g ON l.group_id = g.id
	LEFT JOIN accounts a ON l.teacher_id = a.id
	WHERE l.id = $1
	`

	var (
		lesson     LessonResponse
		date       time.Time
		startTime  time.Time
		endTime    time.Time
		firstName  *string
		middleName *string
		lastName   *string
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.Group.ID,
		&lesson.Group.Name,
		&date,
		&startTime,
		&endTime,
		&lesson.Subject,
		&lesson.Type,
		&lesson.Teacher.ID,
		&firstName,
		&middleName,
		&lastName,
		&lesson.Classroom,
	)
	if err != nil {
		logger.LogSQLError(err, query, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("lesson not found")
		}
		return nil, err
	}

	lesson.Date = date.Format("2006-01-02")
	lesson.StartTime = startTime.Format("15:04")
	lesson.EndTime = endTime.Format("15:04")

	var parts []string
	if lastName != nil && *lastName != "" {
		parts = append(parts, *lastName)
	}
	if firstName != nil && *firstName != "" {
		parts = append(parts, *firstName)
	}
	if middleName != nil && *middleName != "" {
		parts = append(parts, *middleName)
	}
	lesson.Teacher.FullName = strings.Join(parts, " ")

	return &lesson, nil
}
