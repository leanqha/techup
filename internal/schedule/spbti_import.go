package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"techup/internal/apperrors"
	"techup/internal/logger"
	"time"
)

// SPbTIScheduleBaseURL is the public schedule page on the institute site. Per
// group the site serves a server-rendered week at
// "<base>/<groupID>/<week>", where <week> is an absolute week counter. We read
// those pages directly (the AJAX endpoint needs a CSRF token; the GET URL does
// not).
const SPbTIScheduleBaseURL = "https://spbti.ru/studentam/schedule"

// defaultImportWeeks / maxImportWeeks bound how many weeks a single import run
// fetches, starting from the current one. Each week is one HTTP request, so the
// ceiling keeps an admin request from fanning out unboundedly.
const (
	defaultImportWeeks = 2
	maxImportWeeks     = 8
)

// spbtiGroup is one entry of the window.SCH_GROUPS array embedded in the base
// page: the site's internal group id ("i") and its display name ("n").
type spbtiGroup struct {
	ID   int    `json:"i"`
	Name string `json:"n"`
}

// parsedLesson is a single lesson parsed from a rendered schedule week. Times
// are kept as "HH:MM" strings and the teacher as the site's short "Фамилия И. О."
// form; both are resolved later against our own catalog/accounts.
type parsedLesson struct {
	Date      time.Time
	StartTime string
	EndTime   string
	Subject   string
	Type      string
	Classroom string
	Teacher   string
}

var (
	reSchGroups = regexp.MustCompile(`(?s)SCH_GROUPS\s*=\s*(\[.*?\])\s*;`)
	reDataWeek  = regexp.MustCompile(`id="sch-state"[^>]*data-week="(\d+)"`)
	reTemplate  = regexp.MustCompile(`(?s)<template id="sch-server-html">(.*?)</template>`)

	reSchDates    = regexp.MustCompile(`(?s)class="sch-dates"[^>]*>(.*?)</`)
	reSchDayName  = regexp.MustCompile(`(?s)class="sch-day-name"[^>]*>(.*?)</div>`)
	reSchItemTime = regexp.MustCompile(`(?s)class="sch-item-time"[^>]*>(.*?)</div>`)
	reSchItemSubj = regexp.MustCompile(`(?s)class="sch-item-subj"[^>]*>(.*?)</div>`)
	reSchPlace    = regexp.MustCompile(`(?s)class="sch-item-place"[^>]*>(.*?)</span>`)
	reSchTeacher  = regexp.MustCompile(`(?s)class="sch-item-teacher"[^>]*>(.*?)</span>`)
	reAnchor      = regexp.MustCompile(`(?s)<a[^>]*>(.*?)</a>`)
	reStripTags   = regexp.MustCompile(`(?s)<[^>]+>`)
	reWhitespace  = regexp.MustCompile(`\s+`)
	reYear        = regexp.MustCompile(`\b(\d{4})\b`)
	reDayMonth    = regexp.MustCompile(`(\d{1,2})\s+([А-Яа-яёЁ]+)`)
	reTeacherFIO  = regexp.MustCompile(`^(.+?)\s+([А-ЯЁ])\.\s*([А-ЯЁ])\.?$`)
	reTimeRange   = regexp.MustCompile(`(\d{1,2}:\d{2})\s*[-–—]\s*(\d{1,2}:\d{2})`)
)

// genitiveMonths maps the Russian genitive month names used in the day headers
// to their month number.
var genitiveMonths = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

// subjectTypeSuffixes maps the trailing "лекц./практ./лаб." marker the site
// appends to a subject onto our lesson-type catalog names. Anything else falls
// back to "Другое".
var subjectTypeSuffixes = map[string]string{
	"лекц":     "Лекция",
	"лекция":   "Лекция",
	"практ":    "Практика",
	"практика": "Практика",
	"лаб":      "Лабораторная",
	"экз":      "Экзамен",
	"зач":      "Зачет",
	"зачет":    "Зачет",
}

// cleanText strips tags, unescapes entities and collapses whitespace.
func cleanText(s string) string {
	s = reStripTags.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = reWhitespace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// parseSPbTIGroups reads the window.SCH_GROUPS array embedded in the base page.
func parseSPbTIGroups(pageHTML string) ([]spbtiGroup, error) {
	m := reSchGroups.FindStringSubmatch(pageHTML)
	if m == nil {
		return nil, apperrors.Internal("group list not found on schedule page", nil)
	}
	var groups []spbtiGroup
	if err := json.Unmarshal([]byte(m[1]), &groups); err != nil {
		return nil, apperrors.Internal("failed to parse group list", err)
	}
	return groups, nil
}

// findSPbTIGroupID resolves a group display name to the site's internal id.
// Matching is case-insensitive on collapsed whitespace.
func findSPbTIGroupID(groups []spbtiGroup, name string) (int, bool) {
	want := strings.ToLower(strings.Join(strings.Fields(name), " "))
	for _, g := range groups {
		if strings.ToLower(strings.Join(strings.Fields(g.Name), " ")) == want {
			return g.ID, true
		}
	}
	return 0, false
}

// parseDataWeek reads the current absolute week counter from the sch-state node.
func parseDataWeek(pageHTML string) (int, bool) {
	m := reDataWeek.FindStringSubmatch(pageHTML)
	if m == nil {
		return 0, false
	}
	w, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return w, true
}

// splitSubjectType peels the trailing lesson-type marker off a subject and
// returns the clean subject plus the mapped catalog type. Subjects with no
// recognised marker keep their full text and get the "Другое" type.
func splitSubjectType(subject string) (string, string) {
	subject = strings.TrimSpace(subject)
	idx := strings.LastIndex(subject, ",")
	if idx >= 0 {
		marker := strings.ToLower(strings.TrimSpace(subject[idx+1:]))
		marker = strings.TrimRight(marker, ".")
		if t, ok := subjectTypeSuffixes[marker]; ok {
			return strings.TrimSpace(subject[:idx]), t
		}
	}
	return subject, "Другое"
}

// parseScheduleWeek extracts the lessons of a single rendered schedule week.
// The schedule markup lives inside a <template id="sch-server-html"> element;
// dates are built from each day header's day+month and the year(s) in the week
// range (the later year is used for January when a week spans the new year).
func parseScheduleWeek(pageHTML string) []parsedLesson {
	body := pageHTML
	if m := reTemplate.FindStringSubmatch(pageHTML); m != nil {
		body = m[1]
	}

	years := weekYears(body)
	if len(years) == 0 {
		return nil
	}

	// Walk day headers in order; every sch-item up to the next header belongs
	// to the current day. We append a sentinel so the last day's trailing items
	// are captured by the same slice logic.
	headerLocs := reSchDayName.FindAllStringSubmatchIndex(body, -1)
	if len(headerLocs) == 0 {
		return nil
	}

	var lessons []parsedLesson
	for i, loc := range headerLocs {
		dayText := cleanText(body[loc[2]:loc[3]])
		date, ok := dayDate(dayText, years)
		if !ok {
			continue
		}

		segEnd := len(body)
		if i+1 < len(headerLocs) {
			segEnd = headerLocs[i+1][0]
		}
		segment := body[loc[1]:segEnd]

		for _, item := range splitItems(segment) {
			l, ok := parseItem(item, date)
			if ok {
				lessons = append(lessons, l)
			}
		}
	}
	return lessons
}

// weekYears returns the distinct 4-digit years present in the week-range label,
// ascending. Most weeks have one; a week spanning the new year has two.
func weekYears(body string) []int {
	m := reSchDates.FindStringSubmatch(body)
	if m == nil {
		return nil
	}
	label := cleanText(m[1])
	seen := map[int]bool{}
	var years []int
	for _, ym := range reYear.FindAllStringSubmatch(label, -1) {
		y, err := strconv.Atoi(ym[1])
		if err != nil || seen[y] {
			continue
		}
		seen[y] = true
		years = append(years, y)
	}
	sort.Ints(years)
	return years
}

// dayDate turns a day header like "Вторник, 20 октября" into a concrete date.
// When the week straddles the new year, January days take the later year.
func dayDate(dayText string, years []int) (time.Time, bool) {
	m := reDayMonth.FindStringSubmatch(dayText)
	if m == nil {
		return time.Time{}, false
	}
	day, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Time{}, false
	}
	month, ok := genitiveMonths[strings.ToLower(m[2])]
	if !ok {
		return time.Time{}, false
	}

	year := years[0]
	if len(years) > 1 && month == time.January {
		year = years[len(years)-1]
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC), true
}

// splitItems slices a day segment into the HTML of its individual lesson items.
func splitItems(segment string) []string {
	idxs := []int{}
	marker := `class="sch-item"`
	for off := 0; ; {
		p := strings.Index(segment[off:], marker)
		if p < 0 {
			break
		}
		idxs = append(idxs, off+p)
		off += p + len(marker)
	}
	var items []string
	for i, start := range idxs {
		end := len(segment)
		if i+1 < len(idxs) {
			end = idxs[i+1]
		}
		items = append(items, segment[start:end])
	}
	return items
}

// parseItem extracts one lesson from a single sch-item fragment.
func parseItem(item string, date time.Time) (parsedLesson, bool) {
	timeM := reSchItemTime.FindStringSubmatch(item)
	subjM := reSchItemSubj.FindStringSubmatch(item)
	if timeM == nil || subjM == nil {
		return parsedLesson{}, false
	}

	tr := reTimeRange.FindStringSubmatch(cleanText(timeM[1]))
	if tr == nil {
		return parsedLesson{}, false
	}
	start, end := normalizeTime(tr[1]), normalizeTime(tr[2])
	if start == "" || end == "" {
		return parsedLesson{}, false
	}

	subject, lessonType := splitSubjectType(cleanText(subjM[1]))
	if subject == "" {
		return parsedLesson{}, false
	}

	classroom := ""
	if m := reSchPlace.FindStringSubmatch(item); m != nil {
		classroom = cleanText(m[1])
	}

	teacher := ""
	if m := reSchTeacher.FindStringSubmatch(item); m != nil {
		if a := reAnchor.FindStringSubmatch(m[1]); a != nil {
			teacher = cleanText(a[1])
		} else {
			teacher = cleanText(m[1])
		}
	}

	return parsedLesson{
		Date:      date,
		StartTime: start,
		EndTime:   end,
		Subject:   subject,
		Type:      lessonType,
		Classroom: classroom,
		Teacher:   teacher,
	}, true
}

// normalizeTime pads a "H:MM" / "HH:MM" string to a canonical "HH:MM".
func normalizeTime(s string) string {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return ""
	}
	return t.Format("15:04")
}

// fetchPage downloads one page and returns its body. The site rejects the
// default Go user agent, so it presents a browser-like one (as the teacher
// import does).
func fetchPage(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", apperrors.Internal("failed to build request", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TechUpBot/1.0; +https://mytechup.ru)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return "", apperrors.Internal("failed to fetch schedule page", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", apperrors.Internal(fmt.Sprintf("schedule page returned status %d", resp.StatusCode), nil)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return "", apperrors.Internal("failed to read schedule page", err)
	}
	return string(raw), nil
}

// teacherMatcher resolves the site's short "Фамилия И. О." teacher names to our
// account ids. A surname+initials key that matches exactly one teacher account
// resolves; an absent or ambiguous key does not, so the lesson is skipped.
type teacherMatcher struct {
	byKey map[string][]int
}

func newTeacherMatcher(teachers []teacherAccount) *teacherMatcher {
	m := &teacherMatcher{byKey: make(map[string][]int, len(teachers))}
	for _, t := range teachers {
		key := teacherKey(t.LastName, firstRune(t.FirstName), firstRune(t.MiddleName))
		if key == "" {
			continue
		}
		m.byKey[key] = append(m.byKey[key], t.ID)
	}
	return m
}

// resolve returns the account id for a short name, or false when there is no
// unambiguous match.
func (m *teacherMatcher) resolve(shortName string) (int, bool) {
	fio := reTeacherFIO.FindStringSubmatch(strings.TrimSpace(shortName))
	if fio == nil {
		return 0, false
	}
	key := teacherKey(fio[1], []rune(fio[2])[0], []rune(fio[3])[0])
	ids := m.byKey[key]
	if len(ids) != 1 {
		return 0, false
	}
	return ids[0], true
}

// teacherKey builds the lower-cased "surname|и|о" lookup key from a surname and
// first/middle initials.
func teacherKey(last string, firstInitial, middleInitial rune) string {
	last = strings.ToLower(strings.TrimSpace(last))
	if last == "" || firstInitial == 0 {
		return ""
	}
	return last + "|" + strings.ToLower(string(firstInitial)) + "|" + strings.ToLower(string(middleInitial))
}

// firstRune returns the first rune of s, or 0 when s is empty.
func firstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}

// teacherAccount is the minimal teacher identity the matcher needs.
type teacherAccount struct {
	ID         int
	FirstName  string
	MiddleName string
	LastName   string
}

// LoadSPbTISchedule fetches a group's schedule from the public spbti.ru site
// and replaces that group's lessons across the fetched weeks. Teachers are
// matched against existing teacher accounts by surname and initials; lessons
// with no unambiguous match are skipped (run the teacher import first to
// populate the roster). The import is idempotent: each run replaces the lessons
// in the dates it covers.
func (s *Service) LoadSPbTISchedule(ctx context.Context, groupName string, weeks int) (*SPbTIImportResponse, error) {
	groupName = strings.TrimSpace(groupName)
	if groupName == "" {
		return nil, apperrors.InvalidArgument("group is required")
	}
	if weeks <= 0 {
		weeks = defaultImportWeeks
	}
	if weeks > maxImportWeeks {
		weeks = maxImportWeeks
	}

	groupID, err := s.repo.GetGroupIDByName(ctx, groupName)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	base, err := fetchPage(ctx, client, SPbTIScheduleBaseURL)
	if err != nil {
		return nil, err
	}
	siteGroups, err := parseSPbTIGroups(base)
	if err != nil {
		return nil, err
	}
	siteGroupID, ok := findSPbTIGroupID(siteGroups, groupName)
	if !ok {
		return nil, apperrors.NotFound("group not found on spbti.ru: " + groupName)
	}

	// The group's default page reports the current absolute week counter.
	firstPage, err := fetchPage(ctx, client, fmt.Sprintf("%s/%d", SPbTIScheduleBaseURL, siteGroupID))
	if err != nil {
		return nil, err
	}
	startWeek, ok := parseDataWeek(firstPage)
	if !ok {
		return nil, apperrors.Internal("could not determine current week from spbti.ru", nil)
	}

	teachers, err := s.repo.GetTeachers(ctx)
	if err != nil {
		return nil, err
	}
	accs := make([]teacherAccount, 0, len(teachers))
	for _, t := range teachers {
		accs = append(accs, teacherAccount{ID: t.ID, FirstName: t.FirstName, MiddleName: t.MiddleName, LastName: t.LastName})
	}
	matcher := newTeacherMatcher(accs)

	var parsed []parsedLesson
	for w := 0; w < weeks; w++ {
		week := startWeek + w
		var pageHTML string
		if w == 0 {
			pageHTML = firstPage // already fetched for startWeek
		} else {
			pageHTML, err = fetchPage(ctx, client, fmt.Sprintf("%s/%d/%d", SPbTIScheduleBaseURL, siteGroupID, week))
			if err != nil {
				return nil, err
			}
		}
		parsed = append(parsed, parseScheduleWeek(pageHTML)...)
	}

	result := &SPbTIImportResponse{Group: groupName, Weeks: weeks, Parsed: len(parsed)}

	var lessons []Lesson
	var minDate, maxDate time.Time
	for _, p := range parsed {
		if minDate.IsZero() || p.Date.Before(minDate) {
			minDate = p.Date
		}
		if maxDate.IsZero() || p.Date.After(maxDate) {
			maxDate = p.Date
		}

		teacherID, ok := matcher.resolve(p.Teacher)
		if !ok {
			result.SkippedNoTeacher++
			continue
		}

		start, err := time.Parse("15:04", p.StartTime)
		if err != nil {
			continue
		}
		end, err := time.Parse("15:04", p.EndTime)
		if err != nil {
			continue
		}

		lessons = append(lessons, Lesson{
			GroupID:   groupID,
			Date:      p.Date,
			StartTime: start,
			EndTime:   end,
			Subject:   p.Subject,
			Type:      p.Type,
			TeacherID: teacherID,
			Classroom: p.Classroom,
		})
	}

	if minDate.IsZero() {
		// Nothing parsed (e.g. empty weeks): leave existing lessons untouched.
		return result, nil
	}

	if err := s.repo.ReplaceGroupLessonsInRange(ctx, groupID, minDate, maxDate, lessons); err != nil {
		return nil, err
	}

	result.Imported = len(lessons)
	result.FromDate = minDate.Format("2006-01-02")
	result.ToDate = maxDate.Format("2006-01-02")

	logger.Log.Info().
		Str("group", groupName).
		Int("weeks", weeks).
		Int("parsed", result.Parsed).
		Int("imported", result.Imported).
		Int("skipped_no_teacher", result.SkippedNoTeacher).
		Msg("imported spbti.ru schedule")

	return result, nil
}
