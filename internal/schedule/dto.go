package schedule

// LessonRequest describes payload for creating or updating a lesson.
type LessonRequest struct {
	Group     int    `json:"group" binding:"required"`
	TeacherID int    `json:"teacher_id"`
	Date      string `json:"date" binding:"required"`       // format: YYYY-MM-DD
	StartTime string `json:"start_time" binding:"required"` // format: HH:MM
	EndTime   string `json:"end_time" binding:"required"`   // format: HH:MM
	Subject   string `json:"subject" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Classroom string `json:"classroom" binding:"required"`
}

// LessonResponse represents a scheduled lesson with related group and teacher info.
type LessonResponse struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Subject   string `json:"subject"`
	Classroom string `json:"classroom"`
	Type      string `json:"type"`

	Group   GroupDTO   `json:"group"`
	Teacher TeacherDTO `json:"teacher"`
}

// FacultyDTO represents a faculty summary.
type FacultyDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GroupDTO represents a student group summary.
type GroupDTO struct {
	ID             int    `json:"id"`
	FacultyID      int    `json:"faculty_id"`
	Name           string `json:"name"`
	Course         int    `json:"course"`
	Degree         string `json:"degree"`
	YearStart      int    `json:"year_start"`
	Specialization string `json:"specialization"`
	IsActive       bool   `json:"is_active"`
}

// TeacherDTO represents a teacher summary.
type TeacherDTO struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	FullName   string `json:"full_name"`
}

// LessonImportRequest describes payload for importing lessons by group name.
type LessonImportRequest struct {
	Group     string `json:"group" binding:"required"`
	TeacherID int    `json:"teacher_id"`
	Date      string `json:"date" binding:"required"`       // format: YYYY-MM-DD
	StartTime string `json:"start_time" binding:"required"` // format: HH:MM
	EndTime   string `json:"end_time" binding:"required"`   // format: HH:MM
	Subject   string `json:"subject" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Classroom string `json:"classroom" binding:"required"`
}

// NoteTextRequest describes payload for note text used by add/update note handlers.
type NoteTextRequest struct {
	Text string `json:"text" binding:"required"`
}

// SubjectRequest describes payload for creating or updating a subject.
type SubjectRequest struct {
	Name string `json:"name" binding:"required"`
}

// LessonTypeRequest describes payload for creating or updating a lesson type.
type LessonTypeRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// GroupSubjectRequest describes payload for adding or updating a curriculum entry.
type GroupSubjectRequest struct {
	GroupID   int  `json:"group_id" binding:"required"`
	SubjectID int  `json:"subject_id" binding:"required"`
	Semester  int  `json:"semester" binding:"required"`
	TeacherID *int `json:"teacher_id"`
}

// GroupSubjectResponse represents a curriculum entry with related display names.
type GroupSubjectResponse struct {
	ID          int    `json:"id"`
	GroupID     int    `json:"group_id"`
	GroupName   string `json:"group_name"`
	SubjectID   int    `json:"subject_id"`
	SubjectName string `json:"subject_name"`
	Semester    int    `json:"semester"`
	TeacherID   *int   `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
}

// SlotRequest describes payload for creating or updating a schedule slot.
type SlotRequest struct {
	GroupID    int    `json:"group_id" binding:"required"`
	Semester   int    `json:"semester" binding:"required"`
	WeekParity string `json:"week_parity" binding:"required"` // "odd" | "even"
	Weekday    int    `json:"weekday" binding:"required"`     // 1=Пн .. 7=Вс
	StartTime  string `json:"start_time" binding:"required"`  // format: HH:MM
	EndTime    string `json:"end_time" binding:"required"`    // format: HH:MM
	Subject    string `json:"subject" binding:"required"`
	Type       string `json:"type"`
	TeacherID  *int   `json:"teacher_id"`
	Classroom  string `json:"classroom"`
}

// SlotResponse represents a schedule slot with related display names.
type SlotResponse struct {
	ID          int    `json:"id"`
	GroupID     int    `json:"group_id"`
	GroupName   string `json:"group_name"`
	Semester    int    `json:"semester"`
	WeekParity  string `json:"week_parity"`
	Weekday     int    `json:"weekday"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Subject     string `json:"subject"`
	Type        string `json:"type"`
	TeacherID   *int   `json:"teacher_id"`
	TeacherName string `json:"teacher_name"`
	Classroom   string `json:"classroom"`
}

// GenerateScheduleRequest describes payload for expanding a group's slots into
// concrete dated lessons across a semester date range.
type GenerateScheduleRequest struct {
	GroupID         int    `json:"group_id" binding:"required"`
	Semester        int    `json:"semester" binding:"required"`
	StartDate       string `json:"start_date" binding:"required"` // format: YYYY-MM-DD
	EndDate         string `json:"end_date" binding:"required"`   // format: YYYY-MM-DD
	FirstWeekParity string `json:"first_week_parity"`             // "odd" | "even"; default "odd"
}

// GenerateScheduleResponse summarizes how many lessons a generation produced.
type GenerateScheduleResponse struct {
	Generated int `json:"generated"`
	Skipped   int `json:"skipped"` // slots skipped because no teacher was assigned
}
