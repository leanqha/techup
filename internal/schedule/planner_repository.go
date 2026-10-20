package schedule

import (
	"context"
	"fmt"
	"strings"
	"techup/internal/apperrors"
	"techup/internal/logger"
	"time"
)

// joinTeacherName builds "Фамилия Имя Отчество" from nullable name columns,
// skipping the parts that are empty or absent.
func joinTeacherName(last, first, middle *string) string {
	var parts []string
	if last != nil && *last != "" {
		parts = append(parts, *last)
	}
	if first != nil && *first != "" {
		parts = append(parts, *first)
	}
	if middle != nil && *middle != "" {
		parts = append(parts, *middle)
	}
	return strings.Join(parts, " ")
}

// ---- Curriculum (group_subjects) ----

func (r *Repository) AddGroupSubject(ctx context.Context, gs *GroupSubject) error {
	query := `
	INSERT INTO group_subjects (group_id, subject_id, semester, teacher_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID).
		Scan(&gs.ID, &gs.CreatedAt, &gs.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already in this group's semester plan")
		}
		logger.LogSQLError(err, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID)
		return err
	}
	return nil
}

func (r *Repository) UpdateGroupSubject(ctx context.Context, gs GroupSubject) error {
	query := `
	UPDATE group_subjects
	SET group_id=$1, subject_id=$2, semester=$3, teacher_id=$4, updated_at=now()
	WHERE id=$5`

	ct, err := r.db.Exec(ctx, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID, gs.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return apperrors.Conflict("subject already in this group's semester plan")
		}
		logger.LogSQLError(err, query, gs.GroupID, gs.SubjectID, gs.Semester, gs.TeacherID, gs.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("curriculum entry not found")
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
		return apperrors.NotFound("curriculum entry not found")
	}
	return nil
}

func (r *Repository) ListGroupSubjects(ctx context.Context, f GroupSubjectFilter) ([]GroupSubjectResponse, error) {
	query := `
	SELECT gs.id, gs.group_id, g.name, gs.subject_id, s.name, gs.semester,
	       gs.teacher_id, a.first_name, a.middle_name, a.last_name
	FROM group_subjects gs
	LEFT JOIN groups g ON gs.group_id = g.id
	LEFT JOIN subjects s ON gs.subject_id = s.id
	LEFT JOIN accounts a ON gs.teacher_id = a.id`

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

	items := make([]GroupSubjectResponse, 0)
	for rows.Next() {
		var (
			dto                    GroupSubjectResponse
			groupName, subjectName *string
			first, middle, last    *string
		)
		if err := rows.Scan(
			&dto.ID, &dto.GroupID, &groupName, &dto.SubjectID, &subjectName, &dto.Semester,
			&dto.TeacherID, &first, &middle, &last,
		); err != nil {
			return nil, err
		}
		if groupName != nil {
			dto.GroupName = *groupName
		}
		if subjectName != nil {
			dto.SubjectName = *subjectName
		}
		dto.TeacherName = joinTeacherName(last, first, middle)
		items = append(items, dto)
	}
	return items, rows.Err()
}

// ---- Schedule slots (schedule_slots) ----

func (r *Repository) AddSlot(ctx context.Context, s *ScheduleSlot) error {
	query := `
	INSERT INTO schedule_slots
		(group_id, semester, week_parity, weekday, start_time, end_time, subject, type, teacher_id, classroom)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		s.GroupID, s.Semester, s.WeekParity, s.Weekday, s.StartTime, s.EndTime,
		s.Subject, s.Type, s.TeacherID, s.Classroom,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		logger.LogSQLError(err, query, s.GroupID, s.Semester, s.WeekParity, s.Weekday)
		return err
	}
	return nil
}

func (r *Repository) UpdateSlot(ctx context.Context, s ScheduleSlot) error {
	query := `
	UPDATE schedule_slots
	SET group_id=$1, semester=$2, week_parity=$3, weekday=$4, start_time=$5, end_time=$6,
		subject=$7, type=$8, teacher_id=$9, classroom=$10, updated_at=now()
	WHERE id=$11`

	ct, err := r.db.Exec(ctx, query,
		s.GroupID, s.Semester, s.WeekParity, s.Weekday, s.StartTime, s.EndTime,
		s.Subject, s.Type, s.TeacherID, s.Classroom, s.ID,
	)
	if err != nil {
		logger.LogSQLError(err, query, s.ID)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("schedule slot not found")
	}
	return nil
}

func (r *Repository) DeleteSlot(ctx context.Context, id int) error {
	query := `DELETE FROM schedule_slots WHERE id=$1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return err
	}
	if ct.RowsAffected() == 0 {
		return apperrors.NotFound("schedule slot not found")
	}
	return nil
}

func (r *Repository) ListSlots(ctx context.Context, f SlotFilter) ([]SlotResponse, error) {
	query := `
	SELECT ss.id, ss.group_id, g.name, ss.semester, ss.week_parity, ss.weekday,
	       ss.start_time, ss.end_time, ss.subject, ss.type, ss.teacher_id,
	       a.first_name, a.middle_name, a.last_name, ss.classroom
	FROM schedule_slots ss
	LEFT JOIN groups g ON ss.group_id = g.id
	LEFT JOIN accounts a ON ss.teacher_id = a.id`

	var conditions []string
	var args []any
	i := 1
	if f.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("ss.group_id = $%d", i))
		args = append(args, *f.GroupID)
		i++
	}
	if f.Semester != nil {
		conditions = append(conditions, fmt.Sprintf("ss.semester = $%d", i))
		args = append(args, *f.Semester)
		i++
	}
	if f.TeacherID != nil {
		conditions = append(conditions, fmt.Sprintf("ss.teacher_id = $%d", i))
		args = append(args, *f.TeacherID)
		i++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY ss.weekday, ss.start_time"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		logger.LogSQLError(err, query, args...)
		return nil, err
	}
	defer rows.Close()

	items := make([]SlotResponse, 0)
	for rows.Next() {
		var (
			dto                 SlotResponse
			groupName           *string
			start, end          time.Time
			first, middle, last *string
		)
		if err := rows.Scan(
			&dto.ID, &dto.GroupID, &groupName, &dto.Semester, &dto.WeekParity, &dto.Weekday,
			&start, &end, &dto.Subject, &dto.Type, &dto.TeacherID,
			&first, &middle, &last, &dto.Classroom,
		); err != nil {
			return nil, err
		}
		if groupName != nil {
			dto.GroupName = *groupName
		}
		dto.StartTime = start.Format("15:04")
		dto.EndTime = end.Format("15:04")
		dto.TeacherName = joinTeacherName(last, first, middle)
		items = append(items, dto)
	}
	return items, rows.Err()
}

// ListSlotsForGeneration returns the raw slots of a group's semester, used to
// expand the template into concrete lessons.
func (r *Repository) ListSlotsForGeneration(ctx context.Context, groupID, semester int) ([]ScheduleSlot, error) {
	query := `
	SELECT id, group_id, semester, week_parity, weekday, start_time, end_time,
	       subject, type, teacher_id, classroom
	FROM schedule_slots
	WHERE group_id = $1 AND semester = $2`

	rows, err := r.db.Query(ctx, query, groupID, semester)
	if err != nil {
		logger.LogSQLError(err, query, groupID, semester)
		return nil, err
	}
	defer rows.Close()

	var slots []ScheduleSlot
	for rows.Next() {
		var s ScheduleSlot
		if err := rows.Scan(
			&s.ID, &s.GroupID, &s.Semester, &s.WeekParity, &s.Weekday, &s.StartTime, &s.EndTime,
			&s.Subject, &s.Type, &s.TeacherID, &s.Classroom,
		); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

// ReplaceGroupLessonsInRange deletes a group's lessons within [from, to] and
// inserts the provided generated lessons, all in a single transaction.
func (r *Repository) ReplaceGroupLessonsInRange(ctx context.Context, groupID int, from, to time.Time, lessons []Lesson) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	deleteQuery := `DELETE FROM lessons WHERE group_id = $1 AND date BETWEEN $2 AND $3`
	if _, err := tx.Exec(ctx, deleteQuery, groupID, from, to); err != nil {
		logger.LogSQLError(err, deleteQuery, groupID, from, to)
		return err
	}

	insertQuery := `
	INSERT INTO lessons (group_id, date, start_time, end_time, subject, type, teacher_id, classroom)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	for _, l := range lessons {
		if _, err := tx.Exec(ctx, insertQuery,
			l.GroupID, l.Date, l.StartTime, l.EndTime, l.Subject, l.Type, l.TeacherID, l.Classroom,
		); err != nil {
			logger.LogSQLError(err, insertQuery, l.GroupID, l.Date)
			return err
		}
	}

	return tx.Commit(ctx)
}
