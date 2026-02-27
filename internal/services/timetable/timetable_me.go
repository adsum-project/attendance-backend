package timetable

import (
	"context"
	"time"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/datetime"
)

func (t *TimetableService) GetOwnTimetable(ctx context.Context, weekStart time.Time) ([]timetablemodels.ClassTimetableItem, error) {
	userID, _ := ctx.Value("userID").(string)
	classes, err := t.repo.GetClassesByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	weekStart = datetime.NormalizeWeekStart(weekStart)
	var result []timetablemodels.ClassTimetableItem
	for _, c := range classes {
		occ := weekStart.AddDate(0, 0, c.DayOfWeek-1)
		modStart, _ := time.Parse("2006-01-02", c.ModuleStartDate)
		modEnd, _ := time.Parse("2006-01-02", c.ModuleEndDate)
		if !occ.Before(modStart) && !occ.After(modEnd) {
			result = append(result, c)
		}
	}

	return result, nil
}
