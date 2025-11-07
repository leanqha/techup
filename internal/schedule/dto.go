package schedule

type UploadScheduleRequest struct {
	FileData []byte `json:"file_data"`
}

// ManualScheduleRequest - payload for manual schedule input
type ManualScheduleRequest struct {
	GroupID int         `json:"group_id" example:"1"`
	Lessons []LessonDTO `json:"lessons"`
}

type LessonDTO struct {
	GroupID    int    `json:"group_id"`
	DayOfWeek  string `json:"day_of_week"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Subject    string `json:"subject"`
	Teacher    string `json:"teacher"`
	Classroom  string `json:"classroom"`
	IsOnline   bool   `json:"is_online"`
	IsEvenWeek bool   `json:"is_even_week"`
}

type ScheduleDTO struct {
	GroupID int         `json:"group_id"`
	Lessons []LessonDTO `json:"lessons"`
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
