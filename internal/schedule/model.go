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

// Lesson — конкретная пара в конкретную дату
type Lesson struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Subject   string    `json:"subject"`
	Teacher   string    `json:"teacher"`
	Classroom string    `json:"classroom"`
	CreatedAt time.Time `json:"created_at"`
}

// LessonNote — персональная заметка пользователя к паре
type LessonNote struct {
	ID        int       `json:"id"`
	UserID    int       `json:"-"`
	LessonID  int       `json:"lesson_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
