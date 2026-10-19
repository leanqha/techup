package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"techup/internal/apperrors"
	"techup/internal/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SPbTITeachersURL is the public "Персональный состав педагогических работников"
// page (Сведения об образовательной организации) that lists the teaching staff.
const SPbTITeachersURL = "https://spbti.ru/sveden/employees/pps/index.html"

// ParsedTeacher is a single teacher parsed from the SPbTI employees page.
type ParsedTeacher struct {
	LastName   string
	FirstName  string
	MiddleName string
}

// FullName joins the name parts back into "Фамилия Имя Отчество".
func (p ParsedTeacher) FullName() string {
	parts := make([]string, 0, 3)
	for _, s := range []string{p.LastName, p.FirstName, p.MiddleName} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " ")
}

// key is the normalised identity used to match a parsed teacher against an
// existing account: lower-cased full name with collapsed whitespace.
func (p ParsedTeacher) key() string {
	return strings.ToLower(strings.Join(strings.Fields(p.FullName()), " "))
}

// TeacherImportResult summarises what an import run changed.
type TeacherImportResult struct {
	Fetched int `json:"fetched"` // teachers parsed from the page
	Created int `json:"created"` // new accounts appended
	Updated int `json:"updated"` // existing accounts whose name fields changed
	Skipped int `json:"skipped"` // already present and unchanged
}

var (
	reTbody  = regexp.MustCompile(`(?s)<tbody[^>]*>(.*?)</tbody>`)
	reRow    = regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	reCell   = regexp.MustCompile(`(?s)<t[dh][^>]*>(.*?)</t[dh]>`)
	reTag    = regexp.MustCompile(`(?s)<[^>]+>`)
	reSpace  = regexp.MustCompile(`\s+`)
	reLetter = regexp.MustCompile(`\p{L}`)
)

// ParseTeachers extracts the teacher roster from the employees page HTML.
//
// The page follows the federally mandated "sveden" layout: a single table
// whose second column ("Фамилия, имя, отчество") holds the full name. We read
// the <tbody> rows and take that column from each, which keeps the parser
// resilient to the many extra columns (degree, qualifications, etc.).
func ParseTeachers(htmlDoc string) []ParsedTeacher {
	body := htmlDoc
	if m := reTbody.FindStringSubmatch(htmlDoc); m != nil {
		body = m[1]
	}

	var teachers []ParsedTeacher
	for _, rowMatch := range reRow.FindAllStringSubmatch(body, -1) {
		cells := reCell.FindAllStringSubmatch(rowMatch[1], -1)
		if len(cells) < 2 {
			continue
		}

		fio := cleanCell(cells[1][1])
		if !looksLikeName(fio) {
			continue
		}

		last, first, middle := splitFIO(fio)
		if last == "" {
			continue
		}
		teachers = append(teachers, ParsedTeacher{
			LastName:   last,
			FirstName:  first,
			MiddleName: middle,
		})
	}
	return teachers
}

// cleanCell strips tags, unescapes entities and collapses whitespace.
func cleanCell(s string) string {
	s = reTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = reSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// looksLikeName filters out placeholder / structural rows that carry no person.
func looksLikeName(s string) bool {
	if s == "" {
		return false
	}
	return reLetter.MatchString(s)
}

// splitFIO splits "Фамилия Имя Отчество" into its parts. Compound patronymics
// ("Шовкет оглы") are kept together as the middle name.
func splitFIO(fio string) (last, first, middle string) {
	fields := strings.Fields(fio)
	switch len(fields) {
	case 0:
		return "", "", ""
	case 1:
		return fields[0], "", ""
	case 2:
		return fields[0], fields[1], ""
	default:
		return fields[0], fields[1], strings.Join(fields[2:], " ")
	}
}

// FetchTeachers downloads the employees page and parses the roster.
func FetchTeachers(ctx context.Context, client *http.Client, url string) ([]ParsedTeacher, error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, apperrors.Internal("failed to build request", err)
	}
	// The site rejects the default Go user agent, so present a browser-like one.
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TechUpBot/1.0; +https://mytechup.ru)")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return nil, apperrors.Internal("failed to fetch employees page", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.Internal(fmt.Sprintf("employees page returned status %d", resp.StatusCode), nil)
	}

	// Cap the read; the page is a couple of MB, so 32 MB is a generous ceiling.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, apperrors.Internal("failed to read employees page", err)
	}

	teachers := ParseTeachers(string(raw))
	if len(teachers) == 0 {
		return nil, apperrors.Internal("no teachers parsed from employees page", nil)
	}
	return teachers, nil
}

// importedTeacherEmail builds a stable, non-routable placeholder email for an
// imported teacher. It is derived from the normalised name so re-runs stay
// idempotent, and the reserved .invalid TLD guarantees it can never receive
// mail. Teachers are matched by name, not by this address.
func importedTeacherEmail(t ParsedTeacher) string {
	sum := sha256.Sum256([]byte(t.key()))
	return "t-" + hex.EncodeToString(sum[:])[:16] + "@pps.spbti.invalid"
}

// placeholderPasswordHash returns a valid bcrypt hash of a random secret. The
// secret is discarded, so the account cannot be logged into until its owner
// goes through password reset.
func placeholderPasswordHash() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword(buf, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ImportTeachers fetches the SPbTI roster and syncs it into the accounts table.
//
// Existing teacher accounts are matched by full name and keep their id (lessons
// reference teachers by id, so ids must stay stable); their name fields are
// updated in place when they differ. Teachers not yet present are appended as
// new accounts, so they receive the next ids at the end of the table and no
// existing id shifts. Nothing is ever deleted.
func (s *Service) ImportTeachers(ctx context.Context) (*TeacherImportResult, error) {
	parsed, err := FetchTeachers(ctx, &http.Client{Timeout: 30 * time.Second}, SPbTITeachersURL)
	if err != nil {
		return nil, err
	}

	role := "teacher"
	existing, err := s.repo.ListAccounts(ctx, AdminAccountsFilter{Role: &role})
	if err != nil {
		return nil, err
	}

	byKey := make(map[string]*Account, len(existing))
	for i := range existing {
		k := strings.ToLower(strings.Join(strings.Fields(fullName(&existing[i])), " "))
		if _, ok := byKey[k]; !ok {
			byKey[k] = &existing[i]
		}
	}

	// One placeholder hash per run is enough: the secret is random and unknown,
	// so sharing it across the run's new rows reveals nothing and avoids running
	// bcrypt hundreds of times inside a single request.
	var sharedHash string

	result := &TeacherImportResult{Fetched: len(parsed)}
	seen := make(map[string]bool, len(parsed))

	for _, t := range parsed {
		k := t.key()
		if seen[k] {
			continue // duplicate line on the page
		}
		seen[k] = true

		if acc, ok := byKey[k]; ok {
			if acc.FirstName == t.FirstName && acc.MiddleName == t.MiddleName && acc.LastName == t.LastName {
				result.Skipped++
				continue
			}
			if err := s.repo.UpdateTeacherName(ctx, acc.ID, t.FirstName, t.MiddleName, t.LastName); err != nil {
				return nil, err
			}
			result.Updated++
			continue
		}

		if sharedHash == "" {
			sharedHash, err = placeholderPasswordHash()
			if err != nil {
				return nil, apperrors.Internal("failed to prepare account", err)
			}
		}

		acc := &Account{
			Email:        importedTeacherEmail(t),
			PasswordHash: sharedHash,
			FirstName:    t.FirstName,
			MiddleName:   t.MiddleName,
			LastName:     t.LastName,
			Role:         role,
		}
		if err := s.repo.CreateTeacher(ctx, acc); err != nil {
			// A unique-email clash means the row already exists (e.g. a race
			// with a concurrent run): treat it as already present.
			if apperrors.IsCode(err, apperrors.CodeConflict) {
				result.Skipped++
				continue
			}
			return nil, err
		}
		result.Created++
	}

	logger.Log.Info().
		Int("fetched", result.Fetched).
		Int("created", result.Created).
		Int("updated", result.Updated).
		Int("skipped", result.Skipped).
		Msg("imported SPbTI teachers")

	return result, nil
}

func fullName(a *Account) string {
	parts := make([]string, 0, 3)
	for _, s := range []string{a.LastName, a.FirstName, a.MiddleName} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " ")
}
