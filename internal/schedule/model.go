package schedule

import "time"

// Faculty represents a university faculty
type Faculty struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// Group represents a student group within a faculty
type Group struct {
	ID             int    `db:"id"`
	FacultyID      int    `db:"faculty_id"`
	Name           string `db:"name"`
	Course         int    `db:"course"`
	Degree         string `db:"degree"` // бакалавриат / магистратура
	YearStart      int    `db:"year_start"`
	Specialization string `db:"specialization"`
	IsActive       bool   `db:"is_active"`
}

// Lesson represents a single class session
type Lesson struct {
	ID         int       `db:"id"`
	GroupID    int       `db:"group_id"`
	DayOfWeek  string    `db:"day_of_week"`
	StartTime  string    `db:"start_time"`
	EndTime    string    `db:"end_time"`
	Subject    string    `db:"subject"`
	Teacher    string    `db:"teacher"`
	Classroom  string    `db:"classroom"`
	IsOnline   bool      `db:"is_online"`
	IsEvenWeek bool      `db:"is_even_week"`
	CreatedAt  time.Time `db:"created_at"`
}
