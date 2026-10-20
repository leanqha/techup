package schedule

import "time"

type SearchLessonsFilter struct {
	Date      *time.Time
	TeacherID *int
	GroupID   *int
	Classroom *string
	Subject   *string
}

// GroupSubjectFilter narrows a group's semester plan listing by group and/or
// semester. Both are optional.
type GroupSubjectFilter struct {
	GroupID  *int
	Semester *int
}
