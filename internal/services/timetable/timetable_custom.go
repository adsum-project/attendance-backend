package timetable

import (
	"context"
	"errors"
	"time"

	timetablemodels "github.com/adsum-project/attendance-backend/internal/models/timetable"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	"github.com/adsum-project/attendance-backend/pkg/utils/datetime"
)

// GetOwnTimetable returns the user's classes for the given week.
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

func (t *TimetableService) GetNodeTimetable(ctx context.Context, room string) ([]timetablemodels.ClassTimetableItem, error) {
	now := time.Now().UTC()
	dayOfWeek := datetime.WeekdayToDayOfWeek(now.Weekday())
	currentTime := now.Format("15:04:05")
	rows, err := t.repo.GetClassesByRoom(ctx, room, dayOfWeek, currentTime)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var result []timetablemodels.ClassTimetableItem
	for _, c := range rows {
		if !seen[c.ClassID] {
			seen[c.ClassID] = true
			result = append(result, c)
		}
	}
	return result, nil
}

func (t *TimetableService) GetNodeRoom(ctx context.Context) (string, error) {
	userID, _ := ctx.Value("userID").(string)
	return t.repo.GetNodeRoomByUserID(ctx, userID)
}

func (t *TimetableService) UpdateNodeRoom(ctx context.Context, room string) error {
	userID, _ := ctx.Value("userID").(string)
	return t.repo.UpsertNodeRoom(ctx, userID, room)
}

func (t *TimetableService) DeleteNodeRoom(ctx context.Context) error {
	userID, _ := ctx.Value("userID").(string)
	return t.repo.DeleteNodeRoom(ctx, userID)
}

var (
	ErrStudentNotEnrolled = errors.New("student is not enrolled in this class")
	ErrClassNotRunning    = errors.New("class is not currently running")
)

// CanStudentSignIntoClass checks enrollment and that the class is currently running.
func (t *TimetableService) CanStudentSignIntoClass(ctx context.Context, userID, classID string) error {
	enrolled, err := t.repo.StudentEnrolledInClass(ctx, userID, classID)
	if err != nil {
		return err
	}
	if !enrolled {
		return ErrStudentNotEnrolled
	}

	now := time.Now().UTC()
	dayOfWeek := datetime.WeekdayToDayOfWeek(now.Weekday())
	currentTime := now.Format("15:04:05")

	running, err := t.repo.ClassCurrentlyRunning(ctx, classID, dayOfWeek, currentTime)
	if err != nil {
		return err
	}
	if !running {
		return ErrClassNotRunning
	}

	return nil
}

func (t *TimetableService) GetClassesEndedRecently(ctx context.Context) ([]timetablerepo.ClassEndedItem, error) {
	now := time.Now().UTC()
	windowStart := now.Add(-10 * time.Minute).Format(time.RFC3339)
	nowStr := now.Format(time.RFC3339)
	dayOfWeek := datetime.WeekdayToDayOfWeek(now.Weekday())
	return t.repo.GetClassesEndedRecently(ctx, windowStart, nowStr, dayOfWeek)
}

func (t *TimetableService) GetEnrolledUserIDsForClass(ctx context.Context, classID string) ([]string, error) {
	return t.repo.GetEnrolledUserIDsForClass(ctx, classID)
}
