package schedule

type LessonDTO struct {
	GroupID   int    `json:"group_id"`
	Date      string `json:"date"`       // ISO string
	StartTime string `json:"start_time"` // ISO string
	EndTime   string `json:"end_time"`   // ISO string
	Subject   string `json:"subject"`
	TeacherID int    `json:"teacher_id"`
	Classroom string `json:"classroom"`
}

type FacultyDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AddGroupDTO struct {
	FacultyID      int    `json:"faculty_id"`
	Name           string `json:"name"`
	Course         int    `json:"course"`
	Degree         string `json:"degree"`
	YearStart      int    `json:"year_start"`
	Specialization string `json:"specialization"`
	IsActive       bool   `json:"is_active"`
}

type TeacherScheduleResponse struct {
	Teacher string   `json:"teacher"`
	Lessons []Lesson `json:"lessons"`
}

type ClassroomScheduleResponse struct {
	Classroom string   `json:"classroom"`
	Lessons   []Lesson `json:"lessons"`
}
