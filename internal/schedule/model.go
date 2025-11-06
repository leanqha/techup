package schedule

import "time"

type Faculty struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
	Type string `db:"type"` // например: "очная", "заочная", "очно-заочная"
}

type Program struct {
	ID        int    `db:"id"`
	FacultyID int    `db:"faculty_id"`
	Name      string `db:"name"`
	Degree    string `db:"degree"` // бакалавриат / магистратура
	Course    int    `db:"course"`
}

type Lesson struct {
	ID          int       `db:"id"`
	ProgramID   int       `db:"program_id"`
	DayOfWeek   string    `db:"day_of_week"`
	StartTime   string    `db:"start_time"`
	EndTime     string    `db:"end_time"`
	Subject     string    `db:"subject"`
	Teacher     string    `db:"teacher"`
	Classroom   string    `db:"classroom"`
	IsOnline    bool      `db:"is_online"`
	GroupNumber string    `db:"group_number"`
	CreatedAt   time.Time `db:"created_at"`
	IsEvenWeek  bool      `db:"is_even_week"`
}

type Schedule struct {
	ProgramID int      `db:"program_id"`
	Lessons   []Lesson `json:"lessons"`
}
