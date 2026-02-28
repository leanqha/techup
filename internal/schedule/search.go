package schedule

import "time"

type SearchLessonsFilter struct {
	Date      *time.Time
	TeacherID *int
	GroupID   *int
	Classroom *string
	Subject   *string
}
