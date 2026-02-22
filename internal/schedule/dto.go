package schedule

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

type FacultyDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

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

type TeacherDTO struct {
	ID         int    `json:"id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	FullName   string `json:"full_name"`
}
