package schedule

import (
	"testing"
	"time"
)

// scheduleFixture mirrors the real spbti.ru markup: a sch-state node with the
// current week, the SCH_GROUPS script, and a sch-server-html template holding
// the week bar and day/item blocks.
const scheduleFixture = `
<script>window.SCH_GROUPS = [{"i":245,"n":"1030"},{"i":99,"n":"144"}];</script>
<div id="sch-state" data-base-url="/studentam/schedule" data-teacher="0" data-group="245" data-week="43"></div>
<template id="sch-server-html">
<div class="sch-weekbar">
  <div class="sch-weekbar-mid">
    <span class="sch-dates">19 - 25 октября 2026</span>
    <span class="sch-parity">Четная неделя</span>
  </div>
</div>
<div class="sch-days">
  <div class="sch-day">
    <div class="sch-day-name">
      Вторник, 20 октября                    </div>
    <div class="sch-item">
      <div class="sch-item-time">09:30 - 11:10</div>
      <div class="sch-item-body">
        <div class="sch-item-subj">Технология неорганических веществ: каталитические процессы, лекц.</div>
        <div class="sch-item-info">
          <span class="sch-item-place"><i class="fa fa-thumbtack"></i> ауд. каф. </span>
          <span class="sch-item-teacher"><i class="fa fa-user"></i><a href="/person/vyacheslav_n_naraev" target="_blank">Нараев В. Н.</a><br></span>
        </div>
      </div>
    </div>
    <div class="sch-item">
      <div class="sch-item-time">11:30 - 13:10</div>
      <div class="sch-item-body">
        <div class="sch-item-subj">Информационные технологии, практ.</div>
        <div class="sch-item-info">
          <span class="sch-item-place"><i class="fa fa-thumbtack"></i> Корпус 1 (Главный корпус). Ауд. 206 (этаж 2). </span>
          <span class="sch-item-teacher"><i class="fa fa-user"></i><a href="/person/x">Безвестный П. Р.</a><br></span>
        </div>
      </div>
    </div>
  </div>
  <div class="sch-day">
    <div class="sch-day-name">
      Среда, 21 октября сегодня                    </div>
    <div class="sch-item">
      <div class="sch-item-time">14:00 - 15:40</div>
      <div class="sch-item-body">
        <div class="sch-item-subj">Кинетика процессов, лаб.</div>
        <div class="sch-item-info">
          <span class="sch-item-place"><i class="fa fa-thumbtack"></i> ауд. каф. </span>
          <span class="sch-item-teacher"><i class="fa fa-user"></i><a href="/person/y">Мальцева Н. В.</a><br></span>
        </div>
      </div>
    </div>
  </div>
</div>
</template>
`

func TestParseSPbTIGroups(t *testing.T) {
	groups, err := parseSPbTIGroups(scheduleFixture)
	if err != nil {
		t.Fatalf("parseSPbTIGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d", len(groups))
	}
	id, ok := findSPbTIGroupID(groups, "1030")
	if !ok || id != 245 {
		t.Fatalf("findSPbTIGroupID(1030) = %d, %v; want 245, true", id, ok)
	}
	if _, ok := findSPbTIGroupID(groups, "нет"); ok {
		t.Fatalf("unexpected match for missing group")
	}
}

func TestParseDataWeek(t *testing.T) {
	w, ok := parseDataWeek(scheduleFixture)
	if !ok || w != 43 {
		t.Fatalf("parseDataWeek = %d, %v; want 43, true", w, ok)
	}
	if _, ok := parseDataWeek("<div>no state</div>"); ok {
		t.Fatalf("unexpected week parsed from empty markup")
	}
}

func TestSplitSubjectType(t *testing.T) {
	cases := []struct {
		in       string
		subject  string
		lessonTp string
	}{
		{"Технология неорганических веществ: каталитические процессы, лекц.", "Технология неорганических веществ: каталитические процессы", "Лекция"},
		{"Информационные технологии, практ.", "Информационные технологии", "Практика"},
		{"Кинетика процессов, лаб.", "Кинетика процессов", "Лабораторная"},
		{"Матанализ, экз.", "Матанализ", "Экзамен"},
		{"Физра, зач.", "Физра", "Зачет"},
		{"Безымянный предмет", "Безымянный предмет", "Другое"},
		{"Предмет с запятой, но без типа", "Предмет с запятой, но без типа", "Другое"},
	}
	for _, c := range cases {
		subj, tp := splitSubjectType(c.in)
		if subj != c.subject || tp != c.lessonTp {
			t.Errorf("splitSubjectType(%q) = (%q, %q); want (%q, %q)", c.in, subj, tp, c.subject, c.lessonTp)
		}
	}
}

func TestDayDateYearSelection(t *testing.T) {
	// Single-year week.
	d, ok := dayDate("Вторник, 20 октября", []int{2026})
	if !ok || !d.Equal(time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("dayDate single year = %v, %v", d, ok)
	}
	// New-year week: January day takes the later year, December the earlier.
	jan, _ := dayDate("Понедельник, 1 января", []int{2026, 2027})
	if !jan.Equal(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("January should use later year, got %v", jan)
	}
	dec, _ := dayDate("Среда, 31 декабря", []int{2026, 2027})
	if !dec.Equal(time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("December should use earlier year, got %v", dec)
	}
	// "сегодня" suffix must not break parsing.
	today, ok := dayDate("Среда, 21 октября сегодня", []int{2026})
	if !ok || today.Day() != 21 || today.Month() != time.October {
		t.Fatalf("dayDate with сегодня suffix = %v, %v", today, ok)
	}
}

func TestParseScheduleWeek(t *testing.T) {
	lessons := parseScheduleWeek(scheduleFixture)
	if len(lessons) != 3 {
		t.Fatalf("want 3 lessons, got %d: %+v", len(lessons), lessons)
	}

	first := lessons[0]
	if !first.Date.Equal(time.Date(2026, time.October, 20, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("first date = %v", first.Date)
	}
	if first.StartTime != "09:30" || first.EndTime != "11:10" {
		t.Errorf("first time = %s-%s", first.StartTime, first.EndTime)
	}
	if first.Subject != "Технология неорганических веществ: каталитические процессы" {
		t.Errorf("first subject = %q", first.Subject)
	}
	if first.Type != "Лекция" {
		t.Errorf("first type = %q", first.Type)
	}
	if first.Classroom != "ауд. каф." {
		t.Errorf("first classroom = %q", first.Classroom)
	}
	if first.Teacher != "Нараев В. Н." {
		t.Errorf("first teacher = %q", first.Teacher)
	}

	if lessons[1].Classroom != "Корпус 1 (Главный корпус). Ауд. 206 (этаж 2)." {
		t.Errorf("second classroom = %q", lessons[1].Classroom)
	}

	// Lesson on the second day belongs to 21 октября.
	if !lessons[2].Date.Equal(time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("third date = %v", lessons[2].Date)
	}
}

func TestTeacherMatcher(t *testing.T) {
	accs := []teacherAccount{
		{ID: 1, LastName: "Нараев", FirstName: "Вячеслав", MiddleName: "Николаевич"},
		{ID: 2, LastName: "Мальцева", FirstName: "Наталья", MiddleName: "Владимировна"},
		// Two teachers that collapse to the same surname+initials key: ambiguous.
		{ID: 3, LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович"},
		{ID: 4, LastName: "Иванов", FirstName: "Игорь", MiddleName: "Ильич"},
	}
	m := newTeacherMatcher(accs)

	if id, ok := m.resolve("Нараев В. Н."); !ok || id != 1 {
		t.Errorf("resolve(Нараев В. Н.) = %d, %v; want 1, true", id, ok)
	}
	if id, ok := m.resolve("Мальцева Н. В."); !ok || id != 2 {
		t.Errorf("resolve(Мальцева Н. В.) = %d, %v; want 2, true", id, ok)
	}
	if _, ok := m.resolve("Иванов И. И."); ok {
		t.Errorf("ambiguous surname+initials should not resolve")
	}
	if _, ok := m.resolve("Неизвестный К. Т."); ok {
		t.Errorf("unknown teacher should not resolve")
	}
	if _, ok := m.resolve("мусор"); ok {
		t.Errorf("malformed name should not resolve")
	}
}
