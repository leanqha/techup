package utils

import "time"

func Combine(date time.Time, clock string) time.Time {
	t, err := time.Parse("15:04", clock)
	if err != nil {
		return time.Time{}
	}

	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		t.Hour(),
		t.Minute(),
		0,
		0,
		date.Location(),
	)
}
