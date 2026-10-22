package schedule

import "time"

// Faculty represents a university faculty
type Faculty struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Group represents a student group within a faculty
type Group struct {
	ID             int    `db:"id" json:"id"`
	FacultyID      int    `db:"faculty_id" json:"faculty_id"`
	Name           string `db:"name" json:"name"`
	Course         int    `db:"course" json:"course"`
	Degree         string `db:"degree" json:"degree"` // бакалавриат / магистратура
	YearStart      int    `db:"year_start" json:"year_start"`
	Specialization string `db:"specialization" json:"specialization"`
	IsActive       bool   `db:"is_active" json:"is_active"`
}

// Subject — предмет из справочника предметов
type Subject struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// LessonType — тип занятия из справочника (с цветом для отображения)
type LessonType struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Color     string    `db:"color" json:"color"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Lesson — конкретная пара в конкретную дату
type Lesson struct {
	ID        int       `json:"id"`
	GroupID   int       `json:"group_id"`
	Date      time.Time `json:"date"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Subject   string    `json:"subject"`
	Type      string    `json:"type"`
	TeacherID *int      `json:"teacher_id"` // nil when no teacher is assigned
	Classroom string    `json:"classroom"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LessonCSV struct {
	Date      time.Time
	Group     string
	StartTime time.Time
	EndTime   time.Time
	Subject   string
	Type      string
	TeacherID int
	Classroom string
}

// GroupSubject — запись учебного плана: предмет, который группа изучает в
// конкретном семестре, с опциональным закреплённым преподавателем.
type GroupSubject struct {
	ID           int       `json:"id"`
	GroupID      int       `json:"group_id"`
	SubjectID    int       `json:"subject_id"`
	Semester     int       `json:"semester"`
	TeacherID    *int      `json:"teacher_id"`
	LessonTypeID *int      `json:"lesson_type_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ScheduleSlot — ячейка конструктора расписания: повторяющееся занятие на
// чётной/нечётной неделе в конкретный день недели и время. Шаблон, из которого
// генерируются конкретные пары (Lesson).
type ScheduleSlot struct {
	ID         int       `json:"id"`
	GroupID    int       `json:"group_id"`
	Semester   int       `json:"semester"`
	WeekParity string    `json:"week_parity"` // "odd" | "even"
	Weekday    int       `json:"weekday"`     // 1=Пн .. 7=Вс
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Subject    string    `json:"subject"`
	Type       string    `json:"type"`
	TeacherID  *int      `json:"teacher_id"`
	Classroom  string    `json:"classroom"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// FacultySession — экзаменационная сессия факультета: период с датами начала и
// конца. У одного факультета может быть несколько сессий (зимняя, летняя и т.д.).
type FacultySession struct {
	ID        int       `json:"id"`
	FacultyID int       `json:"faculty_id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Note — персональная заметка пользователя к паре
type Note struct {
	ID        int       `json:"id"`
	UserID    int       `json:"-"`
	LessonID  int       `json:"lesson_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LessonImport struct {
	GroupName string
	Date      time.Time
	StartTime time.Time
	EndTime   time.Time
	Subject   string
	Type      string
	TeacherID int
	Classroom string
}

// optionalTeacherID converts a plain teacher id into the nullable form a Lesson
// stores: a missing teacher (id <= 0) becomes nil so the lesson is written with
// teacher_id NULL rather than a zero id that would violate the foreign key.
func optionalTeacherID(id int) *int {
	if id <= 0 {
		return nil
	}
	return &id
}
