package schedule

import (
	"context"
	"testing"
	"time"

	"techup/internal/apperrors"

	"github.com/stretchr/testify/assert"
)

func tm(t *testing.T, hhmm string) time.Time {
	t.Helper()
	parsed, err := time.Parse("15:04", hhmm)
	assert.NoError(t, err)
	return parsed
}

func day(t *testing.T, ymd string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", ymd)
	assert.NoError(t, err)
	return parsed
}

func intPtr(v int) *int { return &v }

func TestExpandSlotsParityAndWeekday(t *testing.T) {
	// 2026-02-02 is a Monday → week 0 (odd). Range covers 4 calendar weeks.
	start := day(t, "2026-02-02")
	end := day(t, "2026-03-01") // Sunday of week 3

	slots := []ScheduleSlot{
		// Odd week, Monday (weekday 1) → 2026-02-02 and 2026-02-16.
		{GroupID: 1, WeekParity: "odd", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "A", Type: "Лекция", TeacherID: intPtr(11)},
		// Even week, Wednesday (weekday 3) → 2026-02-11 and 2026-02-25.
		{GroupID: 1, WeekParity: "even", Weekday: 3, StartTime: tm(t, "11:00"), EndTime: tm(t, "12:30"), Subject: "B", Type: "Практика", TeacherID: intPtr(12)},
	}

	lessons := expandSlots(slots, start, end, "odd")

	var dates []string
	for _, l := range lessons {
		dates = append(dates, l.Date.Format("2006-01-02"))
	}
	assert.ElementsMatch(t, []string{"2026-02-02", "2026-02-16", "2026-02-11", "2026-02-25"}, dates)
}

func TestExpandSlotsSkipsTeacherlessAndClampsRange(t *testing.T) {
	start := day(t, "2026-02-02")
	end := day(t, "2026-02-08") // only week 0 (odd)

	slots := []ScheduleSlot{
		{GroupID: 1, WeekParity: "odd", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "A", TeacherID: intPtr(11)},
		{GroupID: 1, WeekParity: "odd", Weekday: 2, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "NoTeacher"},                 // skipped: no teacher
		{GroupID: 1, WeekParity: "even", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "C", TeacherID: intPtr(11)}, // even week not in range
	}

	lessons := expandSlots(slots, start, end, "odd")
	assert.Len(t, lessons, 1)
	assert.Equal(t, "A", lessons[0].Subject)
	assert.Equal(t, 11, lessons[0].TeacherID)
}

func TestExpandSlotsFirstWeekEven(t *testing.T) {
	start := day(t, "2026-02-02") // Monday, week 0 — declared even
	end := day(t, "2026-02-08")

	slots := []ScheduleSlot{
		{GroupID: 1, WeekParity: "even", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "Even", TeacherID: intPtr(11)},
		{GroupID: 1, WeekParity: "odd", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "Odd", TeacherID: intPtr(11)},
	}

	lessons := expandSlots(slots, start, end, "even")
	assert.Len(t, lessons, 1)
	assert.Equal(t, "Even", lessons[0].Subject)
}

func TestGenerateScheduleReplacesAndCounts(t *testing.T) {
	var replacedGroup int
	var replacedFrom, replacedTo time.Time
	var inserted []Lesson

	repo := &mockRepo{
		listSlotsForGenerationFn: func(_ context.Context, groupID, semester int) ([]ScheduleSlot, error) {
			return []ScheduleSlot{
				{GroupID: groupID, Semester: semester, WeekParity: "odd", Weekday: 1, StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "A", TeacherID: intPtr(11)},
				{GroupID: groupID, Semester: semester, WeekParity: "odd", Weekday: 2, Subject: "NoTeacher", StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30")},
			}, nil
		},
		replaceGroupLessonsInRangeFn: func(_ context.Context, groupID int, from, to time.Time, lessons []Lesson) error {
			replacedGroup, replacedFrom, replacedTo, inserted = groupID, from, to, lessons
			return nil
		},
	}
	service := NewService(repo)

	res, err := service.GenerateSchedule(context.Background(), 7, 3, day(t, "2026-02-02"), day(t, "2026-02-08"), "odd")
	assert.NoError(t, err)
	assert.Equal(t, 1, res.Generated) // one teacher-backed slot on an odd Monday in range
	assert.Equal(t, 1, res.Skipped)   // one slot without a teacher
	assert.Equal(t, 7, replacedGroup)
	assert.Equal(t, "2026-02-02", replacedFrom.Format("2006-01-02"))
	assert.Equal(t, "2026-02-08", replacedTo.Format("2006-01-02"))
	assert.Len(t, inserted, 1)
}

func TestGenerateScheduleRejectsBadRange(t *testing.T) {
	service := NewService(&mockRepo{})
	_, err := service.GenerateSchedule(context.Background(), 1, 1, day(t, "2026-03-01"), day(t, "2026-02-01"), "odd")
	assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalidArgument))
}

func TestServiceAddSlotValidation(t *testing.T) {
	service := NewService(&mockRepo{})

	// end before start
	err := service.AddSlot(context.Background(), &ScheduleSlot{
		GroupID: 1, Semester: 1, WeekParity: "odd", Weekday: 1,
		StartTime: tm(t, "10:30"), EndTime: tm(t, "09:00"), Subject: "A",
	})
	assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalidArgument))

	// bad parity
	err = service.AddSlot(context.Background(), &ScheduleSlot{
		GroupID: 1, Semester: 1, WeekParity: "weird", Weekday: 1,
		StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "A",
	})
	assert.True(t, apperrors.IsCode(err, apperrors.CodeInvalidArgument))

	// defaults type to "Другое" on a valid slot
	slot := &ScheduleSlot{
		GroupID: 1, Semester: 1, WeekParity: "even", Weekday: 2,
		StartTime: tm(t, "09:00"), EndTime: tm(t, "10:30"), Subject: "A",
	}
	err = service.AddSlot(context.Background(), slot)
	assert.NoError(t, err)
	assert.Equal(t, "Другое", slot.Type)
}
