package schedule

type UploadScheduleRequest struct {
	FileData []byte `json:"file_data"`
}

type LessonDTO struct {
	ProgramID   int    `json:"program_id"`
	DayOfWeek   int    `json:"day_of_week"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Subject     string `json:"subject"`
	Teacher     string `json:"teacher"`
	Classroom   string `json:"classroom"`
	IsOnline    bool   `json:"is_online"`
	GroupNumber int    `json:"group_number"`
	IsEvenWeek  bool   `json:"is_even_week"`
}

type ScheduleDTO struct {
	ProgramID int         `json:"program_id"`
	Lessons   []LessonDTO `json:"lessons"`
}
