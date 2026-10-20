package schedule

import "time"

type SearchLessonsFilter struct {
	Date      *time.Time
	TeacherID *int
	GroupID   *int
	Classroom *string
	Subject   *string
}

// GroupSubjectFilter narrows a curriculum listing to a group and/or semester.
type GroupSubjectFilter struct {
	GroupID  *int
	Semester *int
}

// SlotFilter narrows a schedule-slot listing. TeacherID is used to find a
// teacher's slots across all groups for cross-group conflict detection.
type SlotFilter struct {
	GroupID   *int
	Semester  *int
	TeacherID *int
}
