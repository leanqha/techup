package account

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleEmployeesHTML = `
<html><body>
<table>
<thead><tr><th>№</th><th>Фамилия, имя, отчество</th><th>Должность</th></tr></thead>
<tbody>
<tr><td>1</td><td><p>Аббас Рашад</p></td><td>Ассистент</td></tr>
<tr><td>2</td><td>Абиев Руфат Шовкет оглы</td><td>Заведующий кафедрой</td></tr>
<tr><td>3</td><td>  Авакова&nbsp;Карина   Константиновна </td><td>Ассистент</td></tr>
<tr><td itemprop="noInfo">&mdash;</td><td itemprop="noInfo">&mdash;</td><td>&mdash;</td></tr>
</tbody>
</table>
</body></html>`

func TestParseTeachers_Sample(t *testing.T) {
	got := ParseTeachers(sampleEmployeesHTML)
	require.Len(t, got, 3)

	assert.Equal(t, ParsedTeacher{LastName: "Аббас", FirstName: "Рашад"}, got[0])
	assert.Equal(t, ParsedTeacher{LastName: "Абиев", FirstName: "Руфат", MiddleName: "Шовкет оглы"}, got[1])
	// Entity-encoded nbsp and ragged whitespace are normalised.
	assert.Equal(t, ParsedTeacher{LastName: "Авакова", FirstName: "Карина", MiddleName: "Константиновна"}, got[2])
}

func TestSplitFIO(t *testing.T) {
	cases := []struct {
		in                  string
		last, first, middle string
	}{
		{"Иванов Иван Иванович", "Иванов", "Иван", "Иванович"},
		{"Абиев Руфат Шовкет оглы", "Абиев", "Руфат", "Шовкет оглы"},
		{"Аббас Рашад", "Аббас", "Рашад", ""},
		{"Сидоров", "Сидоров", "", ""},
		{"   ", "", "", ""},
	}
	for _, c := range cases {
		last, first, middle := splitFIO(c.in)
		assert.Equal(t, c.last, last, c.in)
		assert.Equal(t, c.first, first, c.in)
		assert.Equal(t, c.middle, middle, c.in)
	}
}

func TestImportedTeacherEmail_StableAndDistinct(t *testing.T) {
	a := ParsedTeacher{LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович"}
	b := ParsedTeacher{LastName: "Петров", FirstName: "Пётр"}

	assert.Equal(t, importedTeacherEmail(a), importedTeacherEmail(a), "same teacher -> same email")
	assert.NotEqual(t, importedTeacherEmail(a), importedTeacherEmail(b), "different teachers -> different emails")
	assert.Contains(t, importedTeacherEmail(a), "@pps.spbti.invalid")
}

// TestParseTeachers_Fixture parses the real saved page when SPBTI_FIXTURE
// points at it, so the parser can be validated against production markup
// without committing a multi-megabyte fixture. It is skipped by default.
func TestParseTeachers_Fixture(t *testing.T) {
	path := os.Getenv("SPBTI_FIXTURE")
	if path == "" {
		t.Skip("set SPBTI_FIXTURE to the saved employees HTML to run this test")
	}
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	got := ParseTeachers(string(raw))
	t.Logf("parsed %d teachers", len(got))
	require.Greater(t, len(got), 400)
	for _, teacher := range got {
		require.NotEmpty(t, teacher.LastName)
	}
}

// TestFetchTeachers_Live exercises the real HTTP fetch + parse against
// spbti.ru. It makes a network request, so it only runs when SPBTI_LIVE is set.
func TestFetchTeachers_Live(t *testing.T) {
	if os.Getenv("SPBTI_LIVE") == "" {
		t.Skip("set SPBTI_LIVE=1 to fetch the live employees page")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got, err := FetchTeachers(ctx, &http.Client{Timeout: 30 * time.Second}, SPbTITeachersURL)
	require.NoError(t, err)
	t.Logf("fetched %d teachers", len(got))
	require.Greater(t, len(got), 400)
}
