package schedule

type LessonDTO struct {
	GroupID         int    `json:"group_id"`
	Date            string `json:"date"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	Subject         string `json:"subject"`
	TeacherFullName string `json:"teacher_full_name"`
	Classroom       string `json:"classroom"`
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
