package datetime

import "time"

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
