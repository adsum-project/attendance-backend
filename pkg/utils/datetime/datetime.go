package datetime

import "time"

// WeekdayToDayOfWeek converts time.Weekday (0=Sunday, 1=Monday, ...) to day-of-week (1=Monday, ..., 7=Sunday).
func WeekdayToDayOfWeek(weekday time.Weekday) int {
	if weekday == 0 {
		return 7
	}
	return int(weekday)
}

func NormalizeWeekStart(t time.Time) time.Time {
	d := t
	day := d.Weekday()
	if day == 0 {
		day = 7
	}
	diff := int(day) - 1
	d = d.AddDate(0, 0, -diff)
	d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return d
}
